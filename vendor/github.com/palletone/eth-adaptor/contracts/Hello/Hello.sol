pragma solidity^0.4.23;

contract Hello {
    address public owner;
    string public info;

    constructor(address _owner) public {
        owner = _owner;
    }

    function saySomething(string _str) public returns(string) {
        info = _str;
        return info;
    }
    
    function testpraram(int256 a, bool b, string str, address addr, bytes bs, bytes32 bs32) public view returns(int256, bool, string, address, bytes, bytes32) {
        return (a, b, str, addr,bs, bs32);
    }
    
    function testpraram2(uint256 a, bool b, string str, address addr, bytes bs, bytes28 bs32) public view returns(uint256, bool, string, address, bytes, bytes28) {
        return (a, b, str, addr,bs, bs32);
    }
    
    function testpraram3(int256 a, int a256, uint256 au, int8 a8) public view returns(int256, int,uint256,int8) {
        return (a, a256, au, a8);
    }
}