// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {AccessControl} from "@openzeppelin/contracts/access/AccessControl.sol";
import {ReentrancyGuard} from "@openzeppelin/contracts/utils/ReentrancyGuard.sol";

interface IClipMintable {
    function mint(address to, uint256 amount) external;
    function genesisSupply() external view returns (uint256);
    function hardCap() external view returns (uint256);
    function totalSupply() external view returns (uint256);
}

/// @notice Delayed, metric-driven CLIP issuance controller.
///
/// Metrics are snapshots supplied by a separately managed oracle role. A
/// positive weighted growth score is required before an emission can be
/// queued. Queueing and execution are separate transactions, with a delay,
/// so a guardian or timelock can review an emission before minting.
contract AdaptiveMinter is AccessControl, ReentrancyGuard {
    bytes32 public constant METRICS_ROLE = keccak256("METRICS_ROLE");
    bytes32 public constant GUARDIAN_ROLE = keccak256("GUARDIAN_ROLE");

    uint256 public constant BPS = 10_000;
    uint256 public immutable epochDuration;
    uint256 public immutable executionDelay;
    uint256 public immutable minActiveUsers;
    uint256 public immutable minGrowthBps;
    uint256 public immutable maxEpochEmissionBps;
    uint256 public immutable annualEmissionCapBps;
    uint256 public immutable hardCap;
    uint256 public immutable genesisTimestamp;
    IClipMintable public immutable token;

    struct Metrics {
        uint256 activeUsers;
        uint256 verifiedHolders;
        uint256 settledVolume;
        bool recorded;
    }

    struct Proposal {
        uint256 epoch;
        address recipient;
        uint256 amount;
        uint256 executeAfter;
        bool executed;
        bool cancelled;
    }

    mapping(uint256 => Metrics) public metrics;
    mapping(bytes32 => Proposal) public proposals;
    mapping(uint256 => uint256) public annualMinted;
    mapping(uint256 => uint256) public annualQueued;
    mapping(uint256 => uint256) public epochAllocated;
    uint256 public lastRecordedEpoch;
    uint256 public nextProposalNonce;
    uint256 public queuedTotal;
    bool public issuancePaused;

    error InvalidConfig();
    error EpochNotReady();
    error MetricsMissing();
    error MetricsNotGrowing();
    error EmissionTooLarge();
    error ProposalExists();
    error ProposalNotReady();
    error IssuanceIsPaused();
    error InvalidRecipient();

    event MetricsRecorded(uint256 indexed epoch, uint256 activeUsers, uint256 verifiedHolders, uint256 settledVolume);
    event EmissionQueued(bytes32 indexed proposalId, uint256 indexed epoch, address indexed recipient, uint256 amount, uint256 executeAfter);
    event EmissionExecuted(bytes32 indexed proposalId, uint256 indexed epoch, address indexed recipient, uint256 amount);
    event EmissionCancelled(bytes32 indexed proposalId);
    event IssuancePauseChanged(bool paused);

    constructor(
        address admin,
        address tokenAddress,
        uint256 maxSupply,
        uint256 epochDurationSeconds,
        uint256 delaySeconds,
        uint256 minimumActiveUsers,
        uint256 minimumGrowthBps,
        uint256 epochCapBps,
        uint256 annualCapBps
    ) {
        if (admin == address(0) || tokenAddress == address(0) || epochDurationSeconds == 0 || epochDurationSeconds > 365 days || delaySeconds == 0) revert InvalidConfig();
        if (minimumGrowthBps == 0 || minimumGrowthBps > BPS || epochCapBps == 0 || annualCapBps < epochCapBps) revert InvalidConfig();
        uint256 genesis = IClipMintable(tokenAddress).genesisSupply();
        if (maxSupply < genesis || maxSupply > IClipMintable(tokenAddress).hardCap()) revert InvalidConfig();
        _grantRole(DEFAULT_ADMIN_ROLE, admin);
        _grantRole(GUARDIAN_ROLE, admin);
        token = IClipMintable(tokenAddress);
        hardCap = maxSupply;
        epochDuration = epochDurationSeconds;
        executionDelay = delaySeconds;
        minActiveUsers = minimumActiveUsers;
        minGrowthBps = minimumGrowthBps;
        maxEpochEmissionBps = epochCapBps;
        annualEmissionCapBps = annualCapBps;
        genesisTimestamp = block.timestamp;
    }

    function currentEpoch() public view returns (uint256) {
        return (block.timestamp - genesisTimestamp) / epochDuration;
    }

    function recordMetrics(uint256 activeUsers, uint256 verifiedHolders, uint256 settledVolume) external onlyRole(METRICS_ROLE) {
        uint256 epoch = currentEpoch();
        if (epoch == 0 || epoch <= lastRecordedEpoch) revert EpochNotReady();
        if (activeUsers < minActiveUsers) revert MetricsNotGrowing();
        metrics[epoch] = Metrics(activeUsers, verifiedHolders, settledVolume, true);
        lastRecordedEpoch = epoch;
        emit MetricsRecorded(epoch, activeUsers, verifiedHolders, settledVolume);
    }

    function emissionFor(uint256 epoch) public view returns (uint256 amount, uint256 growthBps) {
        if (epoch == 0 || !metrics[epoch].recorded || !metrics[epoch - 1].recorded) revert MetricsMissing();
        Metrics memory previous = metrics[epoch - 1];
        Metrics memory latest = metrics[epoch];
        uint256 usersGrowth = _growth(previous.activeUsers, latest.activeUsers);
        uint256 holdersGrowth = _growth(previous.verifiedHolders, latest.verifiedHolders);
        uint256 volumeGrowth = _growth(previous.settledVolume, latest.settledVolume);
        growthBps = (usersGrowth * 50 + holdersGrowth * 30 + volumeGrowth * 20) / 100;
        if (growthBps < minGrowthBps) revert MetricsNotGrowing();
        if (growthBps > BPS) growthBps = BPS;
        amount = token.genesisSupply() * growthBps * maxEpochEmissionBps / (BPS * BPS);
        uint256 year = _yearForEpoch(epoch);
        uint256 annualCap = token.genesisSupply() * annualEmissionCapBps / BPS;
        if (annualMinted[year] + amount > annualCap) amount = annualCap > annualMinted[year] ? annualCap - annualMinted[year] : 0;
        if (amount > hardCap) amount = hardCap;
    }

    function queueEmission(uint256 epoch, address recipient) external onlyRole(DEFAULT_ADMIN_ROLE) returns (bytes32 proposalId) {
        if (issuancePaused) revert IssuanceIsPaused();
        if (recipient == address(0)) revert InvalidRecipient();
        if (epoch == 0 || epoch > currentEpoch()) revert EpochNotReady();
        (uint256 amount,) = emissionFor(epoch);
        uint256 year = _yearForEpoch(epoch);
        uint256 epochCap = token.genesisSupply() * maxEpochEmissionBps / BPS;
        if (amount == 0 || annualMinted[year] + annualQueued[year] + amount > token.genesisSupply() * annualEmissionCapBps / BPS) revert EmissionTooLarge();
        if (epochAllocated[epoch] + amount > epochCap) revert EmissionTooLarge();
        if (totalIssued() + queuedTotal + amount > hardCap) revert EmissionTooLarge();
        proposalId = keccak256(abi.encode(address(this), epoch, recipient, amount, nextProposalNonce++));
        if (proposals[proposalId].executeAfter != 0) revert ProposalExists();
        proposals[proposalId] = Proposal(epoch, recipient, amount, block.timestamp + executionDelay, false, false);
        annualQueued[year] += amount;
        epochAllocated[epoch] += amount;
        queuedTotal += amount;
        emit EmissionQueued(proposalId, epoch, recipient, amount, block.timestamp + executionDelay);
    }

    function executeEmission(bytes32 proposalId) external nonReentrant {
        Proposal storage proposal = proposals[proposalId];
        if (issuancePaused) revert IssuanceIsPaused();
        if (proposal.executeAfter == 0 || proposal.cancelled || proposal.executed || block.timestamp < proposal.executeAfter) revert ProposalNotReady();
        if (totalIssued() + proposal.amount > hardCap) revert EmissionTooLarge();
        proposal.executed = true;
        uint256 year = _yearForEpoch(proposal.epoch);
        annualQueued[year] -= proposal.amount;
        annualMinted[year] += proposal.amount;
        queuedTotal -= proposal.amount;
        token.mint(proposal.recipient, proposal.amount);
        emit EmissionExecuted(proposalId, proposal.epoch, proposal.recipient, proposal.amount);
    }

    function cancelEmission(bytes32 proposalId) external onlyRole(GUARDIAN_ROLE) {
        Proposal storage proposal = proposals[proposalId];
        if (proposal.executeAfter == 0 || proposal.executed || proposal.cancelled) revert ProposalNotReady();
        proposal.cancelled = true;
        uint256 year = _yearForEpoch(proposal.epoch);
        annualQueued[year] -= proposal.amount;
        epochAllocated[proposal.epoch] -= proposal.amount;
        queuedTotal -= proposal.amount;
        emit EmissionCancelled(proposalId);
    }

    function setIssuancePaused(bool paused) external onlyRole(GUARDIAN_ROLE) {
        issuancePaused = paused;
        emit IssuancePauseChanged(paused);
    }

    function totalIssued() public view returns (uint256) {
        return token.totalSupply();
    }

    function _yearForEpoch(uint256 epoch) private view returns (uint256) {
        return (epoch * epochDuration) / 365 days;
    }

    function _growth(uint256 beforeValue, uint256 afterValue) private pure returns (uint256) {
        if (beforeValue == 0) return afterValue == 0 ? 0 : BPS;
        if (afterValue <= beforeValue) return 0;
        uint256 delta = afterValue - beforeValue;
        if (delta > type(uint256).max / BPS) return BPS;
        uint256 growth = delta * BPS / beforeValue;
        return growth > BPS ? BPS : growth;
    }
}
