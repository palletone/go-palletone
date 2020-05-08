docker build -t palletone/gptn-onenode . --no-cache

docker run -d --name gptn -p 8545:8545 palletone/gptn-onenode

