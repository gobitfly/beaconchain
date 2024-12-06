###############################################################################
###                          Generate ABI                                   ###
###############################################################################
generate-abi: clean ## Generate contract ABI Go code
	solc --evm-version paris --optimize -o bin/contract --bin --abi --overwrite IERC20.sol
	solc --evm-version paris --optimize -o bin/contract --bin --abi --overwrite Token.sol
	solc --evm-version paris --optimize -o bin/contract --bin --abi --overwrite Multicall3.sol
	solc --evm-version paris --optimize -o bin/contract --bin --abi --overwrite IMulticall3.sol

	abigen --abi=./bin/contract/IERC20.abi --type IERC20 --pkg=contracts --out=./ierc20.go
	abigen --abi=./bin/contract/Token.abi --bin=./bin/contract/Token.bin -type Token --pkg=contracts --out=./token.go
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