pragma solidity ^0.4.24;

/**
 * @title SafeMath
 * @dev Math operations with safety checks that throw on error
 */
library LSafeMath {

  /**
  * @dev Multiplies two numbers, throws on overflow.
  */
  function mul(uint256 a, uint256 b) internal pure returns (uint256) {
    if (a == 0) {
      return 0;
    }
    uint256 c = a * b;
    require(c / a == b);
    return c;
  }

  /**
  * @dev Integer division of two numbers, truncating the quotient.
  */
  function div(uint256 a, uint256 b) internal pure returns (uint256) {
    require(b > 0); // Solidity automatically throws when dividing by 0
    uint256 c = a / b;
    // assert(a == b * c + a % b); // There is no case in which this doesn't hold
    return c;
  }

  /**
  * @dev Substracts two numbers, throws on overflow (i.e. if subtrahend is greater than minuend).
  */
  function sub(uint256 a, uint256 b) internal pure returns (uint256) {
    require(b <= a);
    return a - b;
  }

  /**
  * @dev Adds two numbers, throws on overflow.
  */
  function add(uint256 a, uint256 b) internal pure returns (uint256) {
    uint256 c = a + b;
    require(c >= a);
    return c;
  }
}

contract ethmultisig {
  using LSafeMath for uint;

  address private admin;//debug   

  address private addrA;
  address private addrB;
  address private addrC;
  address private addrD;
  mapping(string => uint8) private reqHistory; 

  event Deposit(address token, address user, uint amount, string ptnaddr);
  event Withdraw(address token, address user, address recver, uint amount, string reqid, uint confirmvalue, string state);
  
  modifier isAdmin() {//debug
    require(msg.sender == admin);
    _;
  }

  constructor(address addra, address addrb, address addrc, address addrd) public {
    admin = msg.sender;//debug
    addrA = addra;
    addrB = addrb;
    addrC = addrc;
    addrD = addrd;
  }

  function setaddrs(address addra, address addrb, address addrc, address addrd) public isAdmin {//debug
    addrA = addra;
    addrB = addrb;
    addrC = addrc;
    addrD = addrd;
  }

  function setoneconfirm(uint8[] addrconfirms, address addr, address[] owners) private pure {
      for (uint8 i=0; i < 4;i++) {
          if (addr != owners[i]) {
              continue;
          }
          addrconfirms[i] = 1;
          break;
      }
  }
    
  function setallconfirms(uint8[] addrconfirms, bytes32 tranhash, address[] owners, bytes sigstr1, bytes sigstr2, bytes sigstr3) private pure {
    address addr = 0;
    if (sigstr1.length != 0) {
        addr = getaddr(tranhash, sigstr1);
        setoneconfirm(addrconfirms, addr, owners);
    }
    if (sigstr2.length != 0) {
        addr = getaddr(tranhash, sigstr2);
        setoneconfirm(addrconfirms, addr, owners);
    }
    if (sigstr3.length != 0) {
        addr = getaddr(tranhash, sigstr3);
        setoneconfirm(addrconfirms, addr, owners);
    }
  }

  function calconfirm(uint8[] addrconfirms) private pure returns (uint8) {
    uint8[] memory weights = new uint8[](4);
    weights[0] = 1;
    weights[1] = 1;
    weights[2] = 1;
    weights[3] = 1;

    uint8 confirms = 0;
    for (uint8 i=0;i<4;i++) {
        confirms += addrconfirms[i]*weights[i];
    }
    return confirms;
  }

  function getconfirm(address[] owners, bytes32 tranhash, bytes sigstr1, bytes sigstr2, bytes sigstr3) private pure returns (uint8)  {
    uint8[] memory addrconfirms = new uint8[](4);

    setallconfirms(addrconfirms, tranhash, owners, sigstr1, sigstr2, sigstr3);

    uint8 confirms = 0;
    confirms = calconfirm(addrconfirms);
    return confirms;  
  }

  function deposit(string ptnaddr) public payable {
    emit Deposit(0, msg.sender, msg.value, ptnaddr);
  }

  function withdraw(address recver, uint amount, string reqid, bytes sigstr1, bytes sigstr2, bytes sigstr3) public {
    require(reqHistory[reqid] == 0);
    
    address[] memory owners = new address[](4);
    owners[0] = addrA;
    owners[1] = addrB;
    owners[2] = addrC;
    owners[3] = addrD;

    uint8 confirms = 0;
    bytes32 tranhash = keccak256(abi.encodePacked(address(this), recver, amount, reqid));//hash
    confirms = getconfirm(owners, tranhash, sigstr1, sigstr2, sigstr3);

    require(confirms >= 3);

    reqHistory[reqid] = 1;
    recver.transfer(amount);
    emit Withdraw(0, msg.sender, recver, amount, reqid, confirms, "withdraw");
  }

  function getmultisig(string reqid) public view returns(uint8) {
    return (reqHistory[reqid]);
  }

  function my_eth_bal() public view returns(uint) {
      return address(this).balance;
  }

  function recover(bytes32 hash, bytes sig) private pure returns (address) {
    bytes32 r;
    bytes32 s;
    uint8 v;

    // Check the signature length
    if (sig.length != 65) {
      return (address(0));
    }

    // Divide the signature in r, s and v variables
    // ecrecover takes the signature parameters, and the only way to get them
    // currently is to use assembly.
    // solium-disable-next-line security/no-inline-assembly
    assembly {
      r := mload(add(sig, 32))
      s := mload(add(sig, 64))
      v := byte(0, mload(add(sig, 96)))
    }

    // Version of signature should be 27 or 28, but 0 and 1 are also possible versions
    if (v < 27) {
      v += 27;
    }

    // If the version is correct return the signer address
    if (v != 27 && v != 28) {
      return (address(0));
    } else {
      // solium-disable-next-line arg-overflow
      return ecrecover(hash, v, r, s);
    }
  }

  function getaddr(bytes32 tranhash, bytes sigstr) private pure returns (address) {
    return recover(tranhash, sigstr);
  }

  function suicideto(address addr) public isAdmin {//debug
      selfdestruct(addr);
  }

  function() payable {
    revert();
  }
}