# Jwt middleware for webapp

## create a signing key

`openssl genrsa -out signkey.rsa 1024` 
`openssl rsa -in signkey.rsa -pubout > verifykey.rsa.pub`