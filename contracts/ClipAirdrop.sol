// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {AccessControl} from "@openzeppelin/contracts/access/AccessControl.sol";
import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import {SafeERC20} from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import {MerkleProof} from "@openzeppelin/contracts/utils/cryptography/MerkleProof.sol";
import {Pausable} from "@openzeppelin/contracts/utils/Pausable.sol";
import {ReentrancyGuard} from "@openzeppelin/contracts/utils/ReentrancyGuard.sol";

/// @notice One-time CLIP claims for wallets proven to hold a particular asset.
/// @dev The asset snapshot and allocation are represented by an immutable
/// Merkle root per campaign. The contract never trusts a wallet address alone.
contract ClipAirdrop is AccessControl, Pausable, ReentrancyGuard {
    using SafeERC20 for IERC20;

    bytes32 public constant CAMPAIGN_ROLE = keccak256("CAMPAIGN_ROLE");
    IERC20 public immutable token;

    struct Campaign {
        bytes32 assetRef;
        bytes32 merkleRoot;
        uint64 start;
        uint64 end;
        uint256 totalAllocated;
        uint256 totalClaimed;
        bool active;
        bool closed;
    }

    mapping(bytes32 => Campaign) public campaigns;
    mapping(bytes32 => mapping(uint256 => uint256)) private claimedBitMap;

    error InvalidCampaign();
    error CampaignExists();
    error CampaignInactive();
    error CampaignExpired();
    error InvalidProof();
    error AlreadyClaimed();
    error AllocationExceeded();
    error TransferFailed();

    event CampaignCreated(bytes32 indexed campaignId, bytes32 indexed assetRef, bytes32 merkleRoot, uint256 totalAllocated, uint64 start, uint64 end);
    event Claimed(bytes32 indexed campaignId, uint256 indexed index, address indexed account, uint256 amount);
    event CampaignPaused(bytes32 indexed campaignId);
    event CampaignClosed(bytes32 indexed campaignId, uint256 swept);

    constructor(address admin, address tokenAddress) {
        if (admin == address(0) || tokenAddress == address(0)) revert InvalidCampaign();
        _grantRole(DEFAULT_ADMIN_ROLE, admin);
        _grantRole(CAMPAIGN_ROLE, admin);
        token = IERC20(tokenAddress);
    }

    function createCampaign(
        bytes32 campaignId,
        bytes32 assetRef,
        bytes32 merkleRoot,
        uint256 totalAllocated,
        uint64 start,
        uint64 end
    ) external onlyRole(CAMPAIGN_ROLE) {
        if (campaignId == bytes32(0) || assetRef == bytes32(0) || merkleRoot == bytes32(0) || totalAllocated == 0 || end <= start || end <= block.timestamp) revert InvalidCampaign();
        if (campaigns[campaignId].merkleRoot != bytes32(0)) revert CampaignExists();
        campaigns[campaignId] = Campaign(assetRef, merkleRoot, start, end, totalAllocated, 0, true, false);
        token.safeTransferFrom(msg.sender, address(this), totalAllocated);
        emit CampaignCreated(campaignId, assetRef, merkleRoot, totalAllocated, start, end);
    }

    function claim(
        bytes32 campaignId,
        uint256 index,
        address account,
        uint256 amount,
        bytes32[] calldata proof
    ) external whenNotPaused nonReentrant {
        Campaign storage campaign = campaigns[campaignId];
        if (!campaign.active) revert CampaignInactive();
        if (block.timestamp < campaign.start) revert CampaignInactive();
        if (block.timestamp > campaign.end) revert CampaignExpired();
        if (account == address(0) || isClaimed(campaignId, index)) revert AlreadyClaimed();
        bytes32 leaf = keccak256(bytes.concat(keccak256(abi.encode(index, account, amount))));
        if (!MerkleProof.verify(proof, campaign.merkleRoot, leaf)) revert InvalidProof();
        if (campaign.totalClaimed + amount > campaign.totalAllocated) revert AllocationExceeded();
        _setClaimed(campaignId, index);
        campaign.totalClaimed += amount;
        token.safeTransfer(account, amount);
        emit Claimed(campaignId, index, account, amount);
    }

    function pauseCampaign(bytes32 campaignId) external onlyRole(CAMPAIGN_ROLE) {
        if (!campaigns[campaignId].active) revert CampaignInactive();
        campaigns[campaignId].active = false;
        emit CampaignPaused(campaignId);
    }

    function pause() external onlyRole(CAMPAIGN_ROLE) {
        _pause();
    }

    function unpause() external onlyRole(CAMPAIGN_ROLE) {
        _unpause();
    }

    function closeCampaign(bytes32 campaignId, address recipient) external onlyRole(CAMPAIGN_ROLE) nonReentrant {
        Campaign storage campaign = campaigns[campaignId];
        if (campaign.merkleRoot == bytes32(0) || campaign.closed || (campaign.active && block.timestamp <= campaign.end)) revert CampaignInactive();
        if (recipient == address(0)) revert InvalidCampaign();
        campaign.active = false;
        campaign.closed = true;
        uint256 remaining = campaign.totalAllocated - campaign.totalClaimed;
        if (remaining > 0) token.safeTransfer(recipient, remaining);
        emit CampaignClosed(campaignId, remaining);
    }

    function isClaimed(bytes32 campaignId, uint256 index) public view returns (bool) {
        uint256 word = claimedBitMap[campaignId][index >> 8];
        uint256 mask = 1 << (index & 255);
        return word & mask == mask;
    }

    function _setClaimed(bytes32 campaignId, uint256 index) private {
        claimedBitMap[campaignId][index >> 8] |= 1 << (index & 255);
    }
}
