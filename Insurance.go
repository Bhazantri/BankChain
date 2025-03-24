// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract InsuranceClaims {
    struct Policy {
        bytes32 policyID;
        address insured;
        uint256 premium;
        uint256 payout;
        bool active;
        bool claimed;
    }

    mapping(bytes32 => Policy) public policies;
    address public oracle;

    event PolicyCreated(bytes32 indexed policyID);
    event ClaimProcessed(bytes32 indexed policyID);

    constructor(address _oracle) {
        oracle = _oracle;
    }

    function createPolicy(bytes32 policyID, uint256 payout) external payable {
        policies[policyID] = Policy(policyID, msg.sender, msg.value, payout, true, false);
        emit PolicyCreated(policyID);
    }

    // Automation: Oracle triggers payout
    function processClaim(bytes32 policyID, bool triggerEvent) external {
        require(msg.sender == oracle, "Only oracle");
        Policy storage policy = policies[policyID];
        require(policy.active && !policy.claimed, "Invalid");

        if (triggerEvent) {
            policy.claimed = true;
            policy.active = false;
            (bool success, ) = policy.insured.call{value: policy.payout}("");
            require(success, "Payout failed");
            emit ClaimProcessed(policyID);
        }
    }
}
