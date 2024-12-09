#! /bin/bash
set -e

clean_up () {
    ARG=$?
    rm -rf /tmp/deposit
    exit $ARG
}
trap clean_up EXIT

# default values from .env file
if [ -f ../.env ]; then
    source ../.env
    bn_endpoint="http://localhost:${CL_PORT}"
    el_endpoint="http://localhost:${EL_PORT}"
fi
index="0"
mnemonic="giant issue aisle success illegal bike spike question tent bar rely arctic volcano long crawl hungry vocal artwork sniff fantasy very lucky have athlete"

while getopts a:b:e:i:m: flag
do
    case "${flag}" in
        b) bn_endpoint=${OPTARG};;
        e) el_endpoint=${OPTARG};;
        i) index=${OPTARG};;
        m) mnemonic=${OPTARG};;
    esac
done

echo "Validator Index: $index";
echo "Mnemonic: $mnemonic";
echo "BN Endpoint: $bn_endpoint";
echo "EL Endpoint: $el_endpoint";

mkdir -p /tmp/deposit
deposit_path="m/44'/60'/0'/0/3"
privatekey="bcdf20249abf0ed6d944c0288fad489e33f66b3960d9e6229c1cd214ed3bbe31"
publickey="0x8943545177806ED17B9F23F0a21ee5948eCaa776"
fork_version=$(curl -s $bn_endpoint/eth/v1/beacon/genesis | jq -r '.data.genesis_fork_version')
deposit_contract_address=$(curl -s $bn_endpoint/eth/v1/config/spec | jq -r '.data.DEPOSIT_CONTRACT_ADDRESS')
eth2-val-tools deposit-data --source-min=192 --source-max=200 --amount=32000000000 --fork-version=$fork_version --withdrawals-mnemonic="$mnemonic" --validators-mnemonic="$mnemonic" > /tmp/deposit/deposits_0-9.txt
while read x; do
    account_name="$(echo "$x" | jq '.account')"
    pubkey="$(echo "$x" | jq '.pubkey')"
    echo "Sending deposit for validator $account_name $pubkey"
    ethereal beacon deposit \
        --allow-unknown-contract=true \
        --address="$deposit_contract_address" \
        --connection=$el_endpoint \
        --data="$x" \
        --value="32000000000" \
        --from="$publickey" \
        --privatekey="$privatekey"
    echo "Sent deposit for validator $account_name $pubkey"
    sleep 3
done < /tmp/deposit/deposits_0-9.txt
exit;
rm -rf /tmp/deposit