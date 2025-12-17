// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/// @notice 模拟 Chainlink CCIP Router，支持 sendMessage 立即调用目标合约的 ccipReceive
contract MockCCIPRouter {
    event Sent(
        address indexed sender,
        address indexed target,
        bytes data,
        uint64 dstChainId
    );

    /// @notice 模拟跨链发送消息
    function sendMessage(
        address target,
        bytes calldata data,
        uint64 dstChainId // 目的链ID，模拟用途
    ) external {
        // 立即调用 target.ccipReceive(data)
        (bool ok, ) = target.call(abi.encodeWithSignature("ccipReceive(bytes)", data));
        require(ok, "target ccipReceive failed");

        emit Sent(msg.sender, target, data, dstChainId);
    }
}
