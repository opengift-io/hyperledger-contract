pragma solidity ^0.4.16;

import "zeppelin-solidity/contracts/token/ERC20/MintableToken.sol";

contract EthGIFTToken is MintableToken {
  string public constant name = "EthGIFTToken";
  string public constant symbol = "EthGIFT";
  uint8 public constant decimals = 18;

  event Burn(address indexed burner, uint256 value);

  function burn() public {
    address burner = msg.sender;
    uint256 _value = balances[burner];
    require(_value>0);

   	balances[burner] = 0;
   	totalSupply_ = totalSupply_.sub(_value);
   	Burn(burner, _value);
  }

}

