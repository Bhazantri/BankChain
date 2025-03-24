// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract EscrowAccounts {
    struct Escrow {
        bytes32 escrowID;
        address buyer;
        address seller;
        uint256 amount;
        bytes32 milestoneHash;
        uint64 deadline;
        bool released;
    }

    mapping(bytes32 => Escrow) public escrows;
    address public keeper;

    event EscrowCreated(bytes32 indexed escrowID);
    event FundsReleased(bytes32 indexed escrowID);

    constructor(address _keeper) {
        keeper = _keeper;
    }

    function createEscrow(bytes32 escrowID, address seller, bytes32 milestoneHash, uint64 duration) external payable {
        escrows[escrowID] = Escrow(escrowID, msg.sender, seller, msg.value, milestoneHash, uint64(block.timestamp + duration), false);
        emit EscrowCreated(escrowID);
    }

    function releaseFunds(bytes32 escrowID, bytes32 milestoneProof) external {
        Escrow storage escrow = escrows[escrowID];
        require(msg.sender == escrow.buyer, "Only buyer");
        require(!escrow.released, "Released");
        require(keccak256(abi.encodePacked(milestoneProof)) == escrow.milestoneHash, "Invalid proof");

        escrow.released = true;
        (bool success, ) = escrow.seller.call{value: escrow.amount}("");
        require(success, "Transfer failed");
        emit FundsReleased(escrowID);
    }

    // Automation: Keeper refunds if deadline passes
    function checkUpkeep(bytes32 escrowID) external view returns (bool) {
        Escrow storage escrow = escrows[escrowID];
        return !escrow.released && block.timestamp > escrow.deadline;
    }

    function performUpkeep(bytes32 escrowID) external {
        require(msg.sender == keeper, "Only keeper");
        Escrow storage escrow = escrows[escrowID];
        if (block.timestamp > escrow.deadline) {
            escrow.released = true;
            (bool success, ) = escrow.buyer.call{value: escrow.amount}("");
            require(success, "Refund failed");
        }
    }
}
