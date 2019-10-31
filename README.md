# value_chain
Symphonyprotocol Value Chain(Alpha Release)

Combined several modules to a simple node.

Log Module:
https://github.com/symphonyprotocol/log
Block Module:
https://github.com/symphonyprotocol/scb
Utils Module:
https://github.com/symphonyprotocol/sutil
P2P Module:
https://github.com/symphonyprotocol/p2p
Wallet and Account Module
https://github.com/symphonyprotocol/swa

## Supported Platforms:
    Linux/MacOS shell
    Windows CMD

## How to deploy
1. Download our `alpha` release and extract it somewhere.
2. **(Optional)** Create a configuration file `~/.symchaincfg` if you want to build your own network:
    ```json
    {
        "nodes":
        [
            {
                "id": "c4ef0694fee0cdf78eab30c83b325293047e0b27511b92e8e206b199b24f13ea",
                "ip":"101.200.156.243",
                "port": 32768
            }
        ]
    }
    ```
    > You can find the `id` in the console log when launching your application for now...
3. Launch the application `main` (`main.exe` for windows) extracted from the release.

## Command line options:
```
commands:
    account
        new             create a new account
        use             use an account as current account for sending transaction or mining
            -addr           wallet addr of the account that will be used as current account
        import          import an account by its wif string
            -wif            wif string
        export          export an account's private key to wif
            -addr           (optional) default: current account
        list            list all accounts
        getbalance      get the balance of an account
            -addr           (optional) default: current account

    transaction
        send            raise a transaction from current or specified account to someone
            -from           (optional) default: current account, the wallet addr where the transaction will be sent from
            -to             the receiver
            -amount         how much SYM will be transferred. 

    blockchain
        new             init a new blockchain
        list            list the content of blockchain

    mine                switch on/off the mining
```

### Example:
![Account Command](https://github.com/symphonyprotocol/value_chain/blob/dev/docs/account.png "Account Command")

## Change logs
### alpha-0.0.0.2:
* used pending pool to resolve the bifurcation
* optimized syncing and mining process
* log utils now works more reliably on windows

## Module Supported
* In progress