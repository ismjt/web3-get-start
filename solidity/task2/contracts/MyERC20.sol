// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/**
 * 作业 1：ERC20 代币
任务：参考 openzeppelin-contracts/contracts/token/ERC20/IERC20.sol实现一个简单的 ERC20 代币合约。要求：
合约包含以下标准 ERC20 功能：
balanceOf：查询账户余额。
transfer：转账。
approve 和 transferFrom：授权和代扣转账。
使用 event 记录转账和授权操作。
提供 mint 函数，允许合约所有者增发代币。
提示：
使用 mapping 存储账户余额和授权信息。
使用 event 定义 Transfer 和 Approval 事件。
部署到sepolia 测试网，导入到自己的钱包
 */

/**
 * @title MyERC20
 * @dev 参考 OpenZeppelin IERC20 接口实现的最小可运行 ERC20 合约
 * 功能：
 *  - 查询余额（balanceOf）
 *  - 代币转账（transfer）
 *  - 授权（approve）和代扣转账（transferFrom）
 *  - 增发（mint，仅限合约所有者）
 *  - 事件记录（Transfer / Approval）
 */
contract MyERC20 {
    // ---------------- 基本信息 ----------------
    uint256 public constant MAX_SUPPLY = 1_000_000_000 * 10 ** 18; // 最大供应量（10亿代币）
    string private _name;
    string private _symbol;
    uint8 public _decimals = 18; // 与 ETH 相同精度

    // ---------------- 状态变量 ----------------
    uint256 public _totalSupply; // 当前代币总量
    address public _owner;       // 合约所有者（可增发）
    mapping(address => uint256) private _balances; // 用户余额表
    // 授权额度表 owner:授权人 spender:被授权人 amount:被授权的代币数量
    mapping(address owner => mapping(address spender => uint amount)) private _allowances;

    // ---------------- 事件定义 ----------------
    event Transfer(address indexed from, address indexed to, uint256 value);
    event Approval(address indexed owner, address indexed spender, uint256 value);

    // ---------------- 修饰器 ----------------
    modifier onlyOwner() {
        require(msg.sender == _owner, "Not contract owner");
        _;
    }

    // ---------------- 构造函数 ----------------
    constructor(string memory name_, string memory symbol_, uint256 initialSupply) {
        _owner = msg.sender;
        _name = name_;
        _symbol = symbol_;

        _mint(_owner, initialSupply * 10 ** uint256(decimals()));
    }

    // ---------------- 核心函数 ----------------

    /// 查询某账户余额
    function balanceOf(address account) public view returns (uint256) {
        return _balances[account];
    }

    /// 普通转账
    function transfer(address to, uint256 amount) public returns (bool) {
        require(to != address(0), "Transfer to zero address");
        require(_balances[msg.sender] >= amount, "Insufficient balance");

        _balances[msg.sender] -= amount;
        _balances[to] += amount;

        emit Transfer(msg.sender, to, amount);
        return true;
    }

    /// 授权某地址可以花费你的代币 TODO：优化 approve 的前后竞争问题（race condition）
    function approve(address spender, uint256 amount) public returns (bool) {
        require(spender != address(0), "Approve to zero address");

        _allowances[msg.sender][spender] = amount;
        emit Approval(msg.sender, spender, amount);
        return true;
    }

    /// 查询授权额度
    function allowance(address tokenOwner, address spender) public view returns (uint256) {
        require(_owner != address(0), "invalid owner address");
        require(spender != address(0), "invalid spender address");
        return _allowances[tokenOwner][spender];
    }

    /// 被授权者代扣转账（transferFrom）
    function transferFrom(address from, address to, uint256 amount) public returns (bool) {
        require(from != address(0) && to != address(0), "Zero address");
        require(_balances[from] >= amount, "Insufficient balance");
        require(_allowances[from][msg.sender] >= amount, "Allowance exceeded");

        _balances[from] -= amount;
        _balances[to] += amount;
        _allowances[from][msg.sender] -= amount;

        emit Transfer(from, to, amount);
        return true;
    }

    /// 增发代币（仅合约拥有者）
    function mint(address to, uint256 amount) public onlyOwner returns (bool) {
        require(totalSupply() + amount <= MAX_SUPPLY, "ERC20: exceeds max supply");
        _mint(to, amount * 10 ** uint256(decimals()));
        return true;
    }

    function totalSupply() public view returns (uint256) {
        return _totalSupply;
    }

    function name() public view returns (string memory) {
        return _name;
    }

    function symbol() public view returns (string memory) {
        return _symbol;
    }

    function decimals() public view returns (uint8) {
        return _decimals;
    }

    // ---------------- 内部函数 ----------------
    function _mint(address to, uint256 amount) internal {
        require(to != address(0), "Mint to zero address");
        _totalSupply += amount;
        _balances[to] += amount;

        emit Transfer(address(0), to, amount);
    }
}
