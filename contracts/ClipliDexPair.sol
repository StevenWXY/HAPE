// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {ERC20} from "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import {SafeERC20} from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import {ReentrancyGuard} from "@openzeppelin/contracts/utils/ReentrancyGuard.sol";

/// @notice Minimal permissionless constant-product pair for BNB/EVM tests.
/// @dev This is not a replacement for audited Uniswap deployments. It is a
/// self-contained pool with explicit fee, deadline and slippage checks.
contract ClipliDexPair is ERC20, ReentrancyGuard {
    using SafeERC20 for IERC20;

    uint256 public constant FEE_DENOMINATOR = 10_000;
    uint256 public constant FEE_BPS = 30;
    uint256 public constant MINIMUM_LIQUIDITY = 1_000;

    IERC20 public immutable token0;
    IERC20 public immutable token1;
    uint112 private reserve0;
    uint112 private reserve1;

    error InvalidTokens();
    error InvalidAmount();
    error InsufficientLiquidity();
    error SlippageExceeded();
    error DeadlineExpired();
    error InvariantViolation();

    event LiquidityAdded(address indexed provider, uint256 amount0, uint256 amount1, uint256 liquidity);
    event LiquidityRemoved(address indexed provider, uint256 amount0, uint256 amount1, uint256 liquidity);
    event Swap(address indexed trader, address indexed tokenIn, uint256 amountIn, uint256 amountOut, address indexed recipient);

    constructor(address tokenA, address tokenB) ERC20("Clipli LP", "CLP-LP") {
        if (tokenA == address(0) || tokenB == address(0) || tokenA == tokenB) revert InvalidTokens();
        if (tokenA < tokenB) {
            token0 = IERC20(tokenA);
            token1 = IERC20(tokenB);
        } else {
            token0 = IERC20(tokenB);
            token1 = IERC20(tokenA);
        }
    }

    function getReserves() external view returns (uint112, uint112) {
        return (reserve0, reserve1);
    }

    function addLiquidity(uint256 amount0, uint256 amount1, uint256 minLiquidity, uint256 deadline)
        external
        nonReentrant
        returns (uint256 liquidity)
    {
        if (block.timestamp > deadline) revert DeadlineExpired();
        if (amount0 == 0 || amount1 == 0) revert InvalidAmount();
        token0.safeTransferFrom(msg.sender, address(this), amount0);
        token1.safeTransferFrom(msg.sender, address(this), amount1);
        if (totalSupply() == 0) {
            liquidity = _sqrt(amount0 * amount1);
            if (liquidity <= MINIMUM_LIQUIDITY) revert InsufficientLiquidity();
            _mint(address(1), MINIMUM_LIQUIDITY);
            liquidity -= MINIMUM_LIQUIDITY;
        } else {
            liquidity = _min(amount0 * totalSupply() / reserve0, amount1 * totalSupply() / reserve1);
        }
        if (liquidity < minLiquidity || liquidity == 0) revert InsufficientLiquidity();
        _mint(msg.sender, liquidity);
        _sync();
        emit LiquidityAdded(msg.sender, amount0, amount1, liquidity);
    }

    function removeLiquidity(uint256 liquidity, uint256 minAmount0, uint256 minAmount1, uint256 deadline)
        external
        nonReentrant
        returns (uint256 amount0, uint256 amount1)
    {
        if (block.timestamp > deadline || liquidity == 0) revert InvalidAmount();
        uint256 supply = totalSupply();
        amount0 = liquidity * reserve0 / supply;
        amount1 = liquidity * reserve1 / supply;
        if (amount0 < minAmount0 || amount1 < minAmount1) revert SlippageExceeded();
        _burn(msg.sender, liquidity);
        token0.safeTransfer(msg.sender, amount0);
        token1.safeTransfer(msg.sender, amount1);
        _sync();
        emit LiquidityRemoved(msg.sender, amount0, amount1, liquidity);
    }

    function swapExactIn(address tokenIn, uint256 amountIn, uint256 minAmountOut, address recipient, uint256 deadline)
        external
        nonReentrant
        returns (uint256 amountOut)
    {
        if (block.timestamp > deadline) revert DeadlineExpired();
        if (recipient == address(0) || amountIn == 0) revert InvalidAmount();
        bool zeroForOne = tokenIn == address(token0);
        if (!zeroForOne && tokenIn != address(token1)) revert InvalidTokens();
        (IERC20 input, IERC20 output, uint256 reserveIn, uint256 reserveOut) = zeroForOne
            ? (token0, token1, reserve0, reserve1)
            : (token1, token0, reserve1, reserve0);
        if (reserveIn == 0 || reserveOut == 0) revert InsufficientLiquidity();
        input.safeTransferFrom(msg.sender, address(this), amountIn);
        uint256 amountInWithFee = amountIn * (FEE_DENOMINATOR - FEE_BPS);
        amountOut = amountInWithFee * reserveOut / (reserveIn * FEE_DENOMINATOR + amountInWithFee);
        if (amountOut < minAmountOut || amountOut == 0 || amountOut >= reserveOut) revert SlippageExceeded();
        output.safeTransfer(recipient, amountOut);
        uint256 balance0 = _balance0();
        uint256 balance1 = _balance1();
        uint256 amount0In = balance0 > reserve0 - (zeroForOne ? 0 : amountOut) ? balance0 - (reserve0 - (zeroForOne ? 0 : amountOut)) : 0;
        uint256 amount1In = balance1 > reserve1 - (zeroForOne ? amountOut : 0) ? balance1 - (reserve1 - (zeroForOne ? amountOut : 0)) : 0;
        uint256 adjusted0 = balance0 * FEE_DENOMINATOR - amount0In * FEE_BPS;
        uint256 adjusted1 = balance1 * FEE_DENOMINATOR - amount1In * FEE_BPS;
        if (adjusted0 * adjusted1 < uint256(reserve0) * reserve1 * FEE_DENOMINATOR * FEE_DENOMINATOR) revert InvariantViolation();
        _sync();
        emit Swap(msg.sender, tokenIn, amountIn, amountOut, recipient);
    }

    function _sync() private {
        uint256 balance0 = _balance0();
        uint256 balance1 = _balance1();
        if (balance0 > type(uint112).max || balance1 > type(uint112).max) revert InvalidAmount();
        reserve0 = uint112(balance0);
        reserve1 = uint112(balance1);
    }

    function _balance0() private view returns (uint256) { return token0.balanceOf(address(this)); }
    function _balance1() private view returns (uint256) { return token1.balanceOf(address(this)); }
    function _min(uint256 a, uint256 b) private pure returns (uint256) { return a < b ? a : b; }
    function _sqrt(uint256 value) private pure returns (uint256 result) {
        if (value == 0) return 0;
        uint256 x = value;
        result = 1;
        if (x >= 2 ** 128) { x >>= 128; result <<= 64; }
        if (x >= 2 ** 64) { x >>= 64; result <<= 32; }
        if (x >= 2 ** 32) { x >>= 32; result <<= 16; }
        if (x >= 2 ** 16) { x >>= 16; result <<= 8; }
        if (x >= 2 ** 8) { x >>= 8; result <<= 4; }
        if (x >= 2 ** 4) { x >>= 4; result <<= 2; }
        if (x >= 2 ** 2) { result <<= 1; }
        for (uint256 i = 0; i < 7; i++) {
            result = (result + value / result) >> 1;
        }
        uint256 roundedDown = value / result;
        return result < roundedDown ? result : roundedDown;
    }
}
