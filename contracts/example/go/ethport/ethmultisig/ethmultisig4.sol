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

contract ERC20 is IERC20 {
    mapping (address => address) public addrmap;
    mapping (address => address) public addrmapPTN;

    event Transfer(address indexed from, address indexed to, uint256 value);
    event Approval(address indexed owner, address indexed spender, uint256 value);

    uint public totalSupply = 10*10**8;
    uint8 constant public decimals = 0;
    string constant public name = "ETH Port";
    string constant public symbol = "ETHPort";

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
        return totalSupply;
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
}

contract ethmultisig is ERC20 {
    using LSafeMath for uint;

    address private admin;//debug

    address private addrA;
    address private addrB;
    address private addrC;
    address private addrD;
    mapping(string => uint8) private reqHistory; 

    event Deposit(address token, address user, uint amount);
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

    function resetMapAddr(address _addr, address _ptnhex) public isAdmin {//debug
        if (addrmap[_addr] == _ptnhex && addrmapPTN[_ptnhex] == _addr) {
            addrmap[_addr] = address(0);
            addrmapPTN[_ptnhex] = address(0);
        } else {
            revert();
        }
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
      emit Deposit(0, msg.sender, msg.value);
    }
}