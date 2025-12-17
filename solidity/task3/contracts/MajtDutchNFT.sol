// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC721/extensions/ERC721URIStorage.sol";
import "@openzeppelin/contracts/access/Ownable.sol";

/**
 * NFT合约
 */
contract MajtDutchNFT is ERC721URIStorage, Ownable {
    uint256 public tokenIdCounter;

    constructor(string memory name, string memory symbol) ERC721(name, symbol) Ownable(msg.sender){
        require(bytes(name).length > 0 && bytes(symbol).length > 0, "Token name and symbol not allow empty");
    }

    function safeMint(address recipient, string memory tokenUriId) public onlyOwner returns (uint256) {
        require(recipient != address(0), "Recipient cannot be zero address");
        require(bytes(tokenUriId).length > 0, "Token URI Id required");

        uint256 newTokenId = tokenIdCounter++;
        _safeMint(recipient, newTokenId);
        _setTokenURI(newTokenId, tokenUriId);

        return newTokenId;
    }

    /// @notice 重写 _baseURI，用于统一 metadata 前缀
    function _baseURI() internal pure override returns (string memory) {
        // 参考：https://github.com/AmazingAng/WTF-Solidity/blob/main/34_ERC721/WTFApe.sol
        return "ipfs://QmeSjSinHpPnmXmspMjwiXyN6zS4E9zccariGR3jxcaWtq/";
    }

}