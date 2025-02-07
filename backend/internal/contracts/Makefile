###############################################################################
###                          Generate ABI                                   ###
###############################################################################
generate-abi: clean ## Generate contract ABI Go code
	solc --evm-version paris --optimize -o bin/contract --bin --abi --overwrite openzeppelin/contracts/token/ERC20/IERC20.sol
	solc --evm-version paris --optimize -o bin/contract --bin --abi --overwrite openzeppelin/contracts/token/ERC20/extensions/IERC20Metadata.sol
	solc --evm-version paris --optimize -o bin/contract --bin --abi --overwrite Token.sol
	solc --evm-version paris --optimize -o bin/contract --bin --abi --overwrite Token721.sol
	solc --evm-version paris --optimize -o bin/contract --bin --abi --overwrite Token1155.sol
	solc --evm-version paris --optimize -o bin/contract --bin --abi --overwrite Multicall3.sol
	solc --evm-version paris --optimize -o bin/contract --bin --abi --overwrite IMulticall3.sol

	abigen --abi=./bin/contract/IERC20.abi --type IERC20 --pkg=contracts --out=./ierc20.go
	abigen --abi=./bin/contract/IERC20Metadata.abi --type IERC20Metadata --pkg=contracts --out=./ierc20metadata.go
	abigen --abi=./bin/contract/Token.abi --bin=./bin/contract/Token.bin -type ERC20 --pkg=contracts --out=./ERC20.go
	abigen --abi=./bin/contract/Token721.abi --bin=./bin/contract/Token721.bin -type ERC721 --pkg=contracts --out=./ERC721.go
	abigen --abi=./bin/contract/Token1155.abi --bin=./bin/contract/Token1155.bin -type ERC1155 --pkg=contracts --out=./ERC1155.go
	abigen --abi=./bin/contract/Multicall3.abi --bin=./bin/contract/Multicall3.bin --type Multicall --pkg=contracts --out=./multicall.go
	abigen --abi=./bin/contract/IMulticall3.abi --type IMulticall3 --pkg=contracts --out=./imulticall3.go

###############################################################################
###                                 Clean                                   ###
###############################################################################
clean:
	@echo "Cleaning..."
	@rm -rf bin

###############################################################################
###                                 Help                                  ###
###############################################################################
help:
	@echo "Available commands:"
	@echo "  make generate-abi  : Generate contract ABI Go code"
	@echo "  make clean      	: Clean the build"
	@echo "  make help       	: Print this help message"
.PHONY: help