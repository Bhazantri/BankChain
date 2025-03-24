// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract TradeFinance {
    struct LetterOfCredit {
        bytes32 lcID;
        address buyer;
        address seller;
        uint256 amount;
        mapping(bytes32 => bool) documents; // BOL hashes
        uint8 docCount;
        bool funded;
        bool completed;
    }

    mapping(bytes32 => LetterOfCredit) public lcs;
    address public oracle;

    event LCCreated(bytes32 indexed lcID);
    event DocumentVerified(bytes32 indexed lcID);

    constructor(address _oracle) {
        oracle = _oracle;
    }

    function createLC(bytes32 lcID, address seller) external payable {
        lcs[lcID] = LetterOfCredit(lcID, msg.sender, seller, msg.value, 0, true, false);
        emit LCCreated(lcID);
    }

    // Automation: Oracle submits verified BOL
    function submitBOL(bytes32 lcID, bytes32 bolHash) external {
        require(msg.sender == oracle, "Only oracle");
        LetterOfCredit storage lc = lcs[lcID];
        require(lc.funded && !lc.completed, "Invalid state");

        if (!lc.documents[bolHash]) {
            lc.documents[bolHash] = true;
            lc.docCount++;
        }

        if (lc.docCount >= 2) {
            lc.completed = true;
            (bool success, ) = lc.seller.call{value: lc.amount}("");
            require(success, "Transfer failed");
        }
        emit DocumentVerified(lcID);
    }
}
