docker build -t palletone/gptn-onenode .

docker run -d --name gptn -p 8545:8545 palletone/gptn-onenode