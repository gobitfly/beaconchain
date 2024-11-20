# Contract ABI

To generate the ABI interface in go you will need to install `solc` and `abigen`.

- solc is available [here](https://docs.soliditylang.org/en/latest/installing-solidity.html)

- and `abigen` can be installed through go directly
```
go install github.com/ethereum/go-ethereum/cmd/abigen@latest
```

You can then run `make generate-abi`.

## Sources and Modification

### ERC20, IERC20, IERC20METADATA, draft-IERC6093, Context
* https://github.com/OpenZeppelin/openzeppelin-contracts/blob/master/contracts/token/ERC20/ERC20.sol
* https://github.com/OpenZeppelin/openzeppelin-contracts/blob/master/contracts/token/ERC20/IERC20.sol
* https://github.com/OpenZeppelin/openzeppelin-contracts/blob/master/contracts/token/ERC20/extensions/IERC20Metadata.sol
* https://github.com/OpenZeppelin/openzeppelin-contracts/blob/master/contracts/interfaces/draft-IERC6093.sol
* https://github.com/OpenZeppelin/openzeppelin-contracts/blob/master/contracts/utils/Context.sol

### Token
* original contract, simple implementation of the abstract contract `ERC20` with a `mint` function

### Multicall3, IMulticall3
* https://github.com/mds1/multicall/blob/main/src/Multicall3.sol
* `IMulticall3` change `payable` functions to `view` to simplify go interaction, it is only suitable for reading