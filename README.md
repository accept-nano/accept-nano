# accept-nano

Payment gateway for [NANO](https://nano.org)

*accept-nano* is a server program that helps you to accept NANO payments easily.

## Installing

 - *accept-nano* is written in Go. You can install from source:
   ```$ go get -u github.com/accept-nano/accept-nano```
   or
 - Download the latest compiled binary from [relases page](https://github.com/accept-nano/accept-nano/releases).

## Running

 - You must have a running NANO node software. It's trivial to setup one. You can find instructions at https://developers.nano.org/guides/node-setup/
 - In NANO node [config](https://github.com/nanocurrency/raiblocks/wiki/config.json), `"rpc_enable"` and `"enable_control"` options must be enabled.
 - Create a config file for *accept-nano*. See [Config section](#config) below.
 - Run it with:
   ```$ accept-nano -config /path/to/the/config.toml```
 - You can run *accept-nano* and NANO node software in the same host but it is not necessary.

## How it works?

 - *accept-nano* is a HTTP server with 2 primary endpoints.
   - **/api/pay** for creating a payment request.
   - **/api/verify** for checking the status of a payment.
 - From client, you create a payment request from client by posting the currency and amount.
 - When *accept-nano* receives a payment request, it creates a random seed and unique address for the payment and saves it in it's database, then returns a unique token to the client.
 - After the payment is created, *accept-nano* starts checking the destination account for incoming funds periodically.
 - While *accept-nano* is checking the payment, the client also checks by calling the verification endpoint. It does this continuously until the payment is verified.
 - Customer has a limited duration to transfer the funds to the destination account. The duration can be set in *accept-nano* config.
 - Then the customer pays the requested amount.
 - If *accept-nano* sees a pending blocks at destination account, it sends a notification to the merchant and changes the status of the payment to "verified".
 - At this point, the payment is received and the merchant is notified. The client can continue it's flow.
 - The server accepts pending blocks at the destination account.
 - The server sends the funds in destination account to the merchant's account defined in the config file.

## Config

 - Config is written in TOML format.
 - The structure of config file is defined in [config.go](https://github.com/accept-nano/accept-nano/blob/master/config.go). See comments for field descriptions.

### Example Config

```toml
EnableDebugLog = false
DatabasePath = "/var/accept-nano.db"
ListenAddress = "0.0.0.0:5000"
NodeURL = "http://192.168.1.100:7076/"
ShutdownTimeout = 5000
RateLimit = "60-H"
Representative = "xrb_1nanode8ngaakzbck8smq6ru9bethqwyehomf79sae1k7xd47dkidjqzffeg"
ReceiveThreshold = "0.000001"
MaxPayments = 10
# Don't forget to set your merchant account.
Account = "xrb_your_mechant_account"
# Generate a new random seed with "accept-nano -seed" command and keep it secret.
Seed = "12F36345AB0B10557F22B36B5FF241EF09AF7AEA00A40B3F52CCD34640040E92"
# NotificationURL is optional.
NotificationURL = "http://localhost:5001/"
AllowedDuration = 3600
# TLS certificate and key if you want to serve HTTPS.
#CertFile = "put.io.crt"
#KeyFile = "put.io.key"
```

## Security

 - *accept-nano* does not need to know your merchant wallet seed. It takes the payments from customers and sends them to your merchant account address defined in config file.
 - *accept-nano* server is designed to be open to the Internet but you can run it in your internal network and control requests to it if you want to be extra safe.
 - *accept-nano* does not keep funds itself and pass incoming payments to the merchant account immediately. So there is only a small period of time when the funds are hold by *accept-nano*.
 - Private keys are not saved in database and derived from the seed you gave in config. So you are safe even if the database file is stolen.

## Contributing

 - Please open an issue if you have a question or suggestion.
 - Don't create a PR before discussing it first.

## Who is using in prodution?

 - [Put.io](https://put.io)
