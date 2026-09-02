// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {AccessControl} from "@openzeppelin/contracts/access/AccessControl.sol";
import {ERC20} from "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import {ERC20Burnable} from "@openzeppelin/contracts/token/ERC20/extensions/ERC20Burnable.sol";
import {ERC20Pausable} from "@openzeppelin/contracts/token/ERC20/extensions/ERC20Pausable.sol";
import {ERC20Permit} from "@openzeppelin/contracts/token/ERC20/extensions/ERC20Permit.sol";

/// @notice CLIP utility token for BNB Smart Chain.
/// @dev Genesis supply is minted once. Further issuance is only possible for
/// the separately governed AdaptiveMinter contract, up to the hard cap.
contract ClipToken is ERC20, ERC20Burnable, ERC20Pausable, ERC20Permit, AccessControl {
    bytes32 public constant MINTER_ROLE = keccak256("MINTER_ROLE");
    bytes32 public constant PAUSER_ROLE = keccak256("PAUSER_ROLE");

    uint256 public immutable genesisSupply;
    uint256 public immutable hardCap;

    error InvalidAddress();
    error InvalidSupply();
    error CapExceeded();

    constructor(
        address admin,
        address treasury,
        uint256 initialSupply,
        uint256 maxSupply
    ) ERC20("Clipli", "CLIP") ERC20Permit("Clipli") {
        if (admin == address(0) || treasury == address(0)) revert InvalidAddress();
        if (initialSupply == 0 || maxSupply < initialSupply) revert InvalidSupply();
        _grantRole(DEFAULT_ADMIN_ROLE, admin);
        _grantRole(PAUSER_ROLE, admin);
        genesisSupply = initialSupply;
        hardCap = maxSupply;
        _mint(treasury, initialSupply);
    }

    function mint(address to, uint256 amount) external onlyRole(MINTER_ROLE) whenNotPaused {
        if (to == address(0)) revert InvalidAddress();
        if (totalSupply() + amount > hardCap) revert CapExceeded();
        _mint(to, amount);
    }

    function pause() external onlyRole(PAUSER_ROLE) {
        _pause();
    }

    function unpause() external onlyRole(PAUSER_ROLE) {
        _unpause();
    }

    function _update(address from, address to, uint256 value)
        internal
        override(ERC20, ERC20Pausable)
    {
        super._update(from, to, value);
    }
}
