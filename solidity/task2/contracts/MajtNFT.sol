// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC721/extensions/ERC721URIStorage.sol";
import "@openzeppelin/contracts/access/Ownable.sol";

contract MajtNFT is ERC721URIStorage, Ownable {
    //记录当前NFT编号
    uint256 public tokenIdCounter;

    //构造函数，设置NFT的名称和符号
    constructor(string memory name, string memory symbol) ERC721(name, symbol) Ownable(msg.sender){
        tokenIdCounter = 0;
    }

    // 铸造NFT函数
    function mintNFT(address recipient, string memory tokenURI) public onlyOwner returns (uint256) {
        uint256 newTokenId = tokenIdCounter;
        _safeMint(recipient, newTokenId);
        _setTokenURI(newTokenId, tokenURI);

        tokenIdCounter += 1;
        return newTokenId;
    }
}