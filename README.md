# accept-nano

[![Build Status](https://travis-ci.org/accept-nano/accept-nano.svg?branch=master)](https://travis-ci.org/accept-nano/accept-nano)

Payment gateway for [NANO](https://nano.org)

*accept-nano* is a server program that helps you to accept NANO payments easily.

## Installing

 - *accept-nano* is written in Go. You can install from source:
   ```$ go get -u github.com/accept-nano/accept-nano```
   or
 - Download the latest compiled binary from [relases page](https://github.com/accept-nano/accept-nano/releases).

## Running

 - You must have a running NANO node. It's trivial to set one up. You can find instructions at https://developers.nano.org/guides/node-setup/
 - In NANO node [config](https://github.com/nanocurrency/raiblocks/wiki/config.json), `"rpc_enable"` and `"enable_control"` options must be enabled.
 - Create a config file for *accept-nano*. See [Config section](#config) below.
 - Run it with:
   ```$ accept-nano -config /path/to/the/config.toml```
 - You can run *accept-nano* and NANO node software in the same host but it is not necessary.

## How it works?

 - *accept-nano* is a HTTP server with 2 primary endpoints.
   - **/api/pay** for creating a payment request.
   - **/api/verify** for checking the status of a payment.
 - From client, you create a payment request by posting the currency and amount.
 - When *accept-nano* receives a payment request, it creates a random seed and unique address for the payment and saves it in its database, then returns a unique token to the client.
 - After the payment is created, *accept-nano* starts checking the destination account for incoming funds periodically.
 - While *accept-nano* is checking the payment, the client also checks by calling the verification endpoint. It does this continuously until the payment is verified.
 - The customer has a limited amount of time to transfer the funds to the destination account. This duration can be set in *accept-nano* config.
 - Then the customer pays the requested amount.
 - If *accept-nano* sees a pending block at destination account, it sends a notification to the merchant and changes the status of the payment to "verified".
 - At this point, the payment is received and the merchant is notified. The client can continue its flow.
 - The server accepts pending blocks at the destination account.
 - The server sends the funds in destination account to the merchants account defined in the config file.

## Config

 - Config is written in TOML format.
 - The structure of config file is defined in [config.go](https://github.com/accept-nano/accept-nano/blob/master/config.go). See comments for field descriptions.

### Example Config

```toml
DatabasePath = "./accept-nano.db"
ListenAddress = "127.0.0.1:8080"
NodeURL = "http://localhost:7076/"
# Don't forget to set your merchant account.
Account = "xrb_your_merchant_account"
# Generate a new random seed with "accept-nano -seed" command and keep it secret.
Seed = "12F36345AB0B10557F22B36B5FF241EF09AF7AEA00A40B3F52CCD34640040E92"
# Payment notifications will be sent to this URL (optional).
NotificationURL = "http://localhost:5000/"
```

## Security

 - *accept-nano* does not need to know your merchant wallet seed. It takes payments from customers and sends them to your merchant account address defined in config file.
 - *accept-nano* server is designed to be open to the Internet but you can run it in your internal network and control requests to it if you want to be extra safe.
 - *accept-nano* does not keep funds itself and passes incoming payments to the merchant account immediately. So there is only a small period of time when the funds are held by *accept-nano*.
 - Private keys are not saved in the database and derived from the seed defined in the config. So you are safe even if the database file is stolen.

## Contributing

 - Please open an issue if you have a question or suggestion.
 - Don't create a PR before discussing it first.

## Who is using *accept-nano* in production?

 - [Put.io](https://put.io)
 - [My Nano Ninja](https://mynano.ninja)

Please send a PR to list your site if *accept-nano* is helping you to receive NANO payments.
