// SPDX-License-Identifier: MIT
pragma solidity ^0.8.21;

/**
 * 创建一个名为Voting的合约，包含以下功能：
一个mapping来存储候选人的得票数
一个vote函数，允许用户投票给某个候选人
一个getVotes函数，返回某个候选人的得票数
一个resetVotes函数，重置所有候选人的得票数
 */

// 自定义error:
// Candidate does not exist
error CandidateInvalidError();
// User had already voted
error RepeatedVotingError();
// Only owner can reset votes
error UnauthorizedError();
// User already voted in this round
error RoundVotedError();

contract Voting {
    event Log(string msg);

    // 合约拥有者
    address public owner;

    // 存储候选人的得票数
    mapping(string => uint256) private votes;

    // 用户对某个候选人的参与投票情况
    mapping(string => string) private userVoteRecord;

    // 用户参与的投票轮次
    mapping(address => uint256) private userVoteRound;
    uint256 private currentRound = 0;

    // 存储所有候选人的名字，用于重置
    string[] private candidates;

    // 构造函数，设置合约拥有者
    constructor(string[] memory _candidates) {
        owner = msg.sender;
        candidates = _candidates;
        // 初始化每个候选人的票数为0
        for (uint i = 0; i < _candidates.length; i++) {
            votes[_candidates[i]] = 0;
        }
    }

    // 投票函数
    function vote(string memory candidate) public {
        // 1、检查候选人是否存在
        bool exists = false;
        for (uint i = 0; i < candidates.length; i++) {
            if (keccak256(bytes(candidates[i])) == keccak256(bytes(candidate))) {
                exists = true;
                break;
            }
        }
        if(!exists){
            revert CandidateInvalidError();
        }

        // 2、检查用户本轮是否已经投票、是否重复投票
        if(userVoteRound[msg.sender] > currentRound){
            revert RoundVotedError();
        }
        string memory key = string(abi.encodePacked(currentRound, ":", msg.sender));
        if(bytes(userVoteRecord[key]).length > 0){
            revert RepeatedVotingError();
        }

        // 3、记录用户投票情况
        userVoteRecord[key] = candidate;
        userVoteRound[msg.sender] = currentRound+1;

        // 4、增加候选人票数
        votes[candidate] += 1;
    }

    // 获取候选人票数
    function getVotes(string memory candidate) public view returns (uint256) {
        return votes[candidate];
    }

    // 重置所有候选人的票数（只有合约拥有者可调用）
    function resetVotes() public {
        if(msg.sender != owner){
            revert UnauthorizedError();
        }
        for (uint i = 0; i < candidates.length; i++) {
            votes[candidates[i]] = 0;
        }
        // 进入下一轮投票，相当于逻辑上清空所有用户投票
        currentRound += 1;

        emit Log(string(abi.encodePacked("resetVotes, current vote round: ", currentRound)));
    }
}