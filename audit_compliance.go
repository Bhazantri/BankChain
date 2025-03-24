// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract AuditCompliance {
    struct Record {
        bytes32 recordID;
        bytes32 dataHash;
        address creator;
        uint64 timestamp;
    }

    mapping(bytes32 => Record) public records;
    mapping(address => bool) public regulators;

    event RecordAdded(bytes32 indexed recordID, bytes32 dataHash);

    constructor(address[] memory _regulators) {
        for (uint i = 0; i < _regulators.length; i++) {
            regulators[_regulators[i]] = true;
        }
    }

    // Automation: Events logged for real-time regulator access
    function addRecord(bytes32 recordID, bytes32 dataHash) external {
        records[recordID] = Record(recordID, dataHash, msg.sender, uint64(block.timestamp));
        emit RecordAdded(recordID, dataHash);
    }

    function getRecord(bytes32 recordID) external view returns (bytes32, address, uint64) {
        require(regulators[msg.sender] || records[recordID].creator == msg.sender, "Unauthorized");
        Record storage rec = records[recordID];
        return (rec.dataHash, rec.creator, rec.timestamp);
    }
}
