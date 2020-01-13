server:
openssl genrsa -out ca.key 2048
openssl req -x509 -new -nodes -key ca.key -subj "/CN=ca.com" -days 5000 -out ca.crt

openssl genrsa -out server.key 2048
openssl req -new -key server.key -subj "/CN=server" -out server.csr
openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt -days 5000

client:
openssl genrsa -out client.key 2048
openssl req -new -key client.key -subj "/CN=client" -out client.csr
openssl x509 -req -in client.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out client.crt -days 5000

/etc/hosts
127.0.0.1       server


refer:
https://blog.csdn.net/u010278923/article/details/79204444
