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


interface IERC20 {
    function transfer(address to, uint256 value) external returns (bool);

    function approve(address spender, uint256 value) external returns (bool);

    function transferFrom(address from, address to, uint256 value) external returns (bool);

    function totalSupply() external view returns (uint256);

    function balanceOf(address who) external view returns (uint256);

    function allowance(address owner, address spender) external view returns (uint256);

    event Transfer(address indexed from, address indexed to, uint256 value);

    event Approval(address indexed owner, address indexed spender, uint256 value);
}


contract PTNMap is IERC20 {
    IERC20 public ptnToken;
    mapping (address => address) public addrmap;
    mapping (address => address) public addrmapPTN;
    //event Deposit(address addr, address ptnhex, uint256 amount);


    event Transfer(address indexed from, address indexed to, uint256 value);
    event Approval(address indexed owner, address indexed spender, uint256 value);

    uint public totalSupply = 10*10**8;
    uint8 constant public decimals = 0;
    string constant public name = "PTN Mapping";
    string constant public symbol = "PTNMap";

    address private admin;

    modifier isAdmin() {//debug
      require(msg.sender == admin);
      _;
    }

    constructor(address _erc20Addr) public {
       admin = msg.sender;
       ptnToken = IERC20(_erc20Addr);
    }

    function transfer(address _ptnhex, uint256 _amt) external returns (bool) {
        if (addrmap[msg.sender] == address(0) && (addrmapPTN[_ptnhex] == address(0))) {
            addrmap[msg.sender] = _ptnhex;
            addrmapPTN[_ptnhex] = msg.sender;
            emit Transfer(msg.sender, _ptnhex, _amt);
            return true;
        } else {
            revert();
        }
    }


    function approve(address spender, uint256 value) external returns (bool){
         return false;
     }

    function transferFrom(address from, address to, uint256 value) external returns (bool){
        return false;
    }

    function totalSupply() external view returns (uint256){
        return ptnToken.totalSupply();
    }

    function allowance(address owner, address spender) external view returns (uint256){
        return 0;
    }

    function balanceOf(address _owner) public view returns (uint) {
        if (addrmap[_owner] == address(0)) {
            if (addrmapPTN[_owner] == address(0)) {
              return 1;
            }
            return 0;
        } else {
            return 0;
        }
    }


    function resetMapAddr(address _addr, address _ptnhex) public isAdmin {
        if (addrmap[_addr] == _ptnhex && addrmapPTN[_ptnhex] == _addr) {
            addrmap[_addr] = address(0);
            addrmapPTN[_ptnhex] = address(0);
        } else {
            revert();
        }
    }


    function getMapPtnAddr(address addr) external view returns (string){
        if (addrmap[addr] == address(0)) {
            return "";
        }
        return encodeBase58(addrmap[addr]);
    }
    function getMapEthAddr(address ptnAddr) external view returns (address){
        return addrmapPTN[ptnAddr];
    }


    function bytesConcat(bytes _b) internal returns (string){
        string memory ret = new string(2 + _b.length);
        bytes memory bret = bytes(ret);
        uint k = 0;
        bret[k++] = byte('P');
        bret[k++] = byte('1');
        for (uint8 i = 0; i < _b.length; i++) bret[k++] = _b[i];
        return string(ret);
    }
    function toBytes(uint160 x) internal returns (bytes) {
        bytes20 a = bytes20(x);
        bytes memory b = new bytes(25);
        b[0] = byte(0);
        for (uint8 i=0; i < 20; i++) {
            b[1+i] = a[i];
        }
        bytes32 cksum = sha256(sha256(0,a));
        for (uint8 j=0; j < 4; j++) {
            b[21+j] = cksum[j];
        }
        return b;
    }
    function encodeBase58(address addrHex) constant returns (string) {
        uint160 a = uint160(addrHex);
        string memory result = bytesConcat(toBase58(toBytes(a)));
        return result;
    }
    function toBase58(bytes source) internal returns (bytes) {
        if (source.length == 0) {
            return "";
        }
        uint8[] memory digits = new uint8[](40); //TODO: figure out exactly how much is needed
        digits[0] = 0;
        uint8 digitlength = 1;
        for (uint8 i = 0; i<source.length; ++i) {
            uint carry = uint8(source[i]);
            for (uint8 j = 0; j<digitlength; ++j) {
                carry += uint(digits[j]) * 256;
                digits[j] = uint8(carry % 58);
                carry = carry / 58;
            }
            
            while (carry > 0) {
                digits[digitlength] = uint8(carry % 58);
                digitlength++;
                carry = carry / 58;
            }
        }

        return toAlphabet(reverse(truncate(digits, digitlength)));
    }
    bytes constant ALPHABET = '123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz';
    function toAlphabet(uint8[] indices) internal returns (bytes) {
        bytes memory output = new bytes(indices.length);
        for (var i = 0; i<indices.length; i++) {
            output[i] = ALPHABET[indices[i]];
        }
        return output;
    }
    function reverse(uint8[] input) internal returns (uint8[]) {
        uint8[] memory output = new uint8[](input.length);
        for (var i = 0; i<input.length; i++) {
            output[i] = input[input.length-1-i];
        }
        return output;
    }
    function truncate(uint8[] array, uint8 length) internal returns (uint8[]) {
        uint8[] memory output = new uint8[](length);
        for (var i = 0; i<length; i++) {
            output[i] = array[i];
        }
        return output;
    }
    function () payable {
        // can receive eth
    }
}
