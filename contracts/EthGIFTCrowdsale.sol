pragma solidity ^0.4.24;

import "openzeppelin-solidity/contracts/token/ERC20/MintableToken.sol";
import "EthGIFTToken.sol";


contract EthGIFTCrowdsale is Ownable{
  using SafeMath for uint256;

  // The token being sold
  EthGIFTToken public token;

  // address where funds are collected
  address public wallet;

  // how many token units a buyer gets per wei
  uint256 public rate;

  // amount of raised money in wei
  uint256 public weiRaised;

  mapping(address => uint256) weiFunds;

  // shutdown sale
  bool public isFinalized = false;

  // withdraw on
  bool public isWithdrawOn = false;

  // min pay
  uint256 public minPayWei;

  // Soft cap
  uint256 public softCapWei;

  // Hard cap
  uint256 public hardCapWei;

  event TokenPurchase(address indexed purchaser, address indexed beneficiary, 
    uint256 value, uint256 amount);


  function EthGIFTCrowdsale(uint256 _minPayWei, uint256 _rate, address _wallet, 
    uint256 _softCapWei,uint256 _hardCapWei) public {
    require(_minPayWei > 0);
    require(_rate > 0);
    require(_wallet != address(0));
    require(_softCapWei > 0);
    require(_hardCapWei >= _softCapWei);

    token = new EthGIFTToken();

    minPayWei = _minPayWei;
    rate = _rate;
    wallet = _wallet;

    softCapWei = _softCapWei;
    hardCapWei = _hardCapWei;
  }


  function finalize() onlyOwner public {
    require(!isFinalized);
    if(weiRaised>=softCapWei){
        // forward raised funds
        wallet.transfer(weiRaised);
      } else {
        isWithdrawOn = true;
      }
    isFinalized = true;
  }


  function () external payable {
    buyTokens(msg.sender);
  }


  function buyTokens(address beneficiary) public payable {
    require(beneficiary != address(0));
    require(msg.value >= minPayWei);

    uint256 weiAmount = msg.value;

    uint256 tokens = weiAmount.mul(rate);

    weiRaised = weiRaised.add(weiAmount);

    weiFunds[beneficiary] = weiFunds[beneficiary].add(weiAmount);

    token.mint(beneficiary, tokens);

    TokenPurchase(msg.sender, beneficiary, weiAmount, tokens);

  }


  function withdraw() public {
    require(isWithdrawOn);
    address burner = msg.sender;
    token.burn();    
    burner.transfer(weiFunds[burner]);
  }
}

