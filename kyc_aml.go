// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract KYCAML {
    struct KYCRecord {
        bytes32 userID;
        bytes32 identityHash;
        bool verified;
        uint64 timestamp;
    }

    mapping(bytes32 => KYCRecord) public records;
    address public oracle;

    event KYCSubmitted(bytes32 indexed userID);
    event KYCVerified(bytes32 indexed userID);

    constructor(address _oracle) {
        oracle = _oracle;
    }

    function submitKYC(bytes32 userID, bytes32 identityHash) external {
        records[userID] = KYCRecord(userID, identityHash, false, uint64(block.timestamp));
        emit KYCSubmitted(userID);
    }

    // Automation: Oracle verifies identity
    function verifyKYC(bytes32 userID) external {
        require(msg.sender == oracle, "Only oracle");
        KYCRecord storage record = records[userID];
        require(!record.verified, "Verified");
        record.verified = true;
        emit KYCVerified(userID);
    }
}
