# Contract ABI

To generate the ABI interface in go you will need to install `solc` and `abigen`.

- solc is available [here](https://docs.soliditylang.org/en/latest/installing-solidity.html)

- and `abigen` can be installed through go directly
```
go install github.com/ethereum/go-ethereum/cmd/abigen@latest
```

Then you need to pull the dependencies contract from OpenZeppelin
```
git clone -b release-v5.2 https://github.com/OpenZeppelin/openzeppelin-contracts.git openzeppelin
```
You can then run `make generate-abi`.

## Sources and Modification

### Tokens
* Token.sol simple implementation of the abstract contract `ERC20` with a `mint` function
* Token721.sol simple implementation of the abstract contract `ERC721` with a `mint` function
* Token1155.sol simple implementation of the abstract contract `ERC1155` with a `mint` function

### Multicall3, IMulticall3
* https://github.com/mds1/multicall/blob/main/src/Multicall3.sol
* `IMulticall3` change `payable` functions to `view` to simplify go interaction, it is only suitable for reading