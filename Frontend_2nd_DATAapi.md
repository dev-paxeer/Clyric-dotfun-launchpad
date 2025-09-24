``` 
curl -X 'GET' \
  'https://paxscan.paxeer.app/api/v2/tokens/0x20896B008A8F25DF672087f5c9BB1Af2A5a02653' \
  -H 'accept: application/json'
  ```

  ## Response

  ```json
  {
  "address_hash": "0x20896B008A8F25DF672087f5c9BB1Af2A5a02653",
  "circulating_market_cap": null,
  "decimals": "18",
  "exchange_rate": null,
  "holders_count": "2",
  "icon_url": null,
  "name": "TestToken",
  "reputation": "ok",
  "symbol": "TST",
  "total_supply": "1000000000000000000000000000",
  "type": "ERC-20",
  "volume_24h": null
}
```

```
curl -X 'GET' \
'https://paxscan.paxeer.app/api/v2/tokens/0x20896B008A8F25DF672087f5c9BB1Af2A5a02653/holders' \
-H 'accept: application/json'
```

## Response

```json 
{
  "items": [
    {
      "address": {
        "ens_domain_name": null,
        "hash": "0x07E8681aAE88CcEE4763D2bDC2A03F498C2e0d52",
        "implementations": [],
        "is_contract": false,
        "is_scam": false,
        "is_verified": false,
        "metadata": null,
        "name": null,
        "private_tags": [],
        "proxy_type": null,
        "public_tags": [],
        "reputation": "ok",
        "watchlist_names": []
      },
      "token_id": null,
      "value": "999950504900014898525046020"
    },
    {
      "address": {
        "ens_domain_name": null,
        "hash": "0x69C9914787dd311dDad1a387B9C120cbd53af68E",
        "implementations": [],
        "is_contract": false,
        "is_scam": false,
        "is_verified": false,
        "metadata": null,
        "name": null,
        "private_tags": [],
        "proxy_type": null,
        "public_tags": [],
        "reputation": "ok",
        "watchlist_names": []
      },
      "token_id": null,
      "value": "49495099985101474953980"
    }
  ],
  "next_page_params": null
}
````


//Get Pool Logs and Swaps

``` 
curl -X 'GET' \
  'https://paxscan.paxeer.app/api/v2/addresses/0x07E8681aAE88CcEE4763D2bDC2A03F498C2e0d52/logs' \
  -H 'accept: application/json'
  ```

  ## Response

  ```json
  {
    "items": [
      {
        "address": {
          "ens_domain_name": null,
          "hash": "0x07E8681aAE88CcEE4763D2bDC2A03F498C2e0d52",
          "implementations": [],
          "is_contract": false,
          "is_scam": false,
          "is_verified": false,
          "metadata": null,
          "name": null,
          "private_tags": [],
          "proxy_type": null,
          "public_tags": [],
          "reputation": "ok",
          "watchlist_names": []
        },
        "block_hash": "0x2ac976834e20db3be927442aa0b9b331578b341f89936e9382d22e05373aa811",
        "block_number": 187373,
        "data": "0x000000000000000000000000000000000000000000000a7b227ec02bcf3afafc00000000000000000000000000000000000000000000000006cd17f086cb54b70000000000000000000000000000000000000000000000000000000000000000",
        "decoded": null,
        "index": 6,
        "smart_contract": {
          "ens_domain_name": null,
          "hash": "0x534DfB04a1A15924daB357694647e4f957543e8F",
          "implementations": [],
          "is_contract": false,
          "is_scam": false,
          "is_verified": false,
          "metadata": null,
          "name": null,
          "private_tags": [],
          "proxy_type": null,
          "public_tags": [],
          "reputation": "ok",
          "watchlist_names": []
        },
        "topics": [
          "0x44ffdb5804b1259013c4234f8f74408561c5ff3b8c4ae37ec6eaabcea38c6db7",
          "0x000000000000000000000000534dfb04a1a15924dab357694647e4f957543e8f",
          "0x00000000000000000000000069c9914787dd311ddad1a387b9c120cbd53af68e",
          null
        ],
        "transaction_hash": "0xf6f8437bda96601cde5716d329054c366764d9762d07f2d0c51666e6ea4b871c"
      },
      {
        "address": {
          "ens_domain_name": null,
          "hash": "0x07E8681aAE88CcEE4763D2bDC2A03F498C2e0d52",
          "implementations": [],
          "is_contract": false,
          "is_scam": false,
          "is_verified": false,
          "metadata": null,
          "name": null,
          "private_tags": [],
          "proxy_type": null,
          "public_tags": [],
          "reputation": "ok",
          "watchlist_names": []
        },
        "block_hash": "0x2ac976834e20db3be927442aa0b9b331578b341f89936e9382d22e05373aa811",
        "block_number": 187373,
        "data": "0x00000000000000000000000000000000000000000000000006de8197f54c90930000000000000000000000000000000000000000033b23c17d51c01118c50504",
        "decoded": {
          "method_call": "Sync(uint256 reserve0, uint256 reserve1)",
          "method_id": "cf2aa508",
          "parameters": [
            {
              "indexed": false,
              "name": "reserve0",
              "type": "uint256",
              "value": "494975498712813715"
            },
            {
              "indexed": false,
              "name": "reserve1",
              "type": "uint256",
              "value": "999950504900014898525046020"
            }
          ]
        },
        "index": 5,
        "smart_contract": {
          "ens_domain_name": null,
          "hash": "0x534DfB04a1A15924daB357694647e4f957543e8F",
          "implementations": [],
          "is_contract": false,
          "is_scam": false,
          "is_verified": false,
          "metadata": null,
          "name": null,
          "private_tags": [],
          "proxy_type": null,
          "public_tags": [],
          "reputation": "ok",
          "watchlist_names": []
        },
        "topics": [
          "0xcf2aa50876cdfbb541206f89af0ee78d44a2abf8d328e37fa4917f982149848a",
          null,
          null,
          null
        ],
        "transaction_hash": "0xf6f8437bda96601cde5716d329054c366764d9762d07f2d0c51666e6ea4b871c"
      },
      {
        "address": {
          "ens_domain_name": null,
          "hash": "0x07E8681aAE88CcEE4763D2bDC2A03F498C2e0d52",
          "implementations": [],
          "is_contract": false,
          "is_scam": false,
          "is_verified": false,
          "metadata": null,
          "name": null,
          "private_tags": [],
          "proxy_type": null,
          "public_tags": [],
          "reputation": "ok",
          "watchlist_names": []
        },
        "block_hash": "0x2ac976834e20db3be927442aa0b9b331578b341f89936e9382d22e05373aa811",
        "block_number": 187373,
        "data": "0x00000000000000000000000000000000000000000000000000000918897473c9000000000000000000000000000000000000000000000000000009184e72a000",
        "decoded": {
          "method_call": "PriceUpdate(uint256 numerator, uint256 denominator)",
          "method_id": "92664190",
          "parameters": [
            {
              "indexed": false,
              "name": "numerator",
              "type": "uint256",
              "value": "10000989975497"
            },
            {
              "indexed": false,
              "name": "denominator",
              "type": "uint256",
              "value": "10000000000000"
            }
          ]
        },
        "index": 4,
        "smart_contract": {
          "ens_domain_name": null,
          "hash": "0x534DfB04a1A15924daB357694647e4f957543e8F",
          "implementations": [],
          "is_contract": false,
          "is_scam": false,
          "is_verified": false,
          "metadata": null,
          "name": null,
          "private_tags": [],
          "proxy_type": null,
          "public_tags": [],
          "reputation": "ok",
          "watchlist_names": []
        },
        "topics": [
          "0x92664190cca12aca9cd5309d87194bdda75bb51362d71c06e1a6f75c7c765711",
          null,
          null,
          null
        ],
        "transaction_hash": "0xf6f8437bda96601cde5716d329054c366764d9762d07f2d0c51666e6ea4b871c"
      },
      {
        "address": {
          "ens_domain_name": null,
          "hash": "0x07E8681aAE88CcEE4763D2bDC2A03F498C2e0d52",
          "implementations": [],
          "is_contract": false,
          "is_scam": false,
          "is_verified": false,
          "metadata": null,
          "name": null,
          "private_tags": [],
          "proxy_type": null,
          "public_tags": [],
          "reputation": "ok",
          "watchlist_names": []
        },
        "block_hash": "0x03a8690bf2c162ead5bd559a89d30f17b82f06d23a4937658381bfaa04439d3e",
        "block_number": 187369,
        "data": "0x0000000000000000000000000000000000000000000000000de0b6b3a76400000000000000000000000000000000000000000000000014f644fd80579e75f5f80000000000000000000000000000000000000000000000000000000000000001",
        "decoded": null,
        "index": 6,
        "smart_contract": {
          "ens_domain_name": null,
          "hash": "0x534DfB04a1A15924daB357694647e4f957543e8F",
          "implementations": [],
          "is_contract": false,
          "is_scam": false,
          "is_verified": false,
          "metadata": null,
          "name": null,
          "private_tags": [],
          "proxy_type": null,
          "public_tags": [],
          "reputation": "ok",
          "watchlist_names": []
        },
        "topics": [
          "0x44ffdb5804b1259013c4234f8f74408561c5ff3b8c4ae37ec6eaabcea38c6db7",
          "0x000000000000000000000000534dfb04a1a15924dab357694647e4f957543e8f",
          "0x00000000000000000000000069c9914787dd311ddad1a387b9c120cbd53af68e",
          null
        ],
        "transaction_hash": "0x52ce00852d100eac200066975fa6bd603e9b65803e59e0a117ce0360650914d3"
      },
      {
        "address": {
          "ens_domain_name": null,
          "hash": "0x07E8681aAE88CcEE4763D2bDC2A03F498C2e0d52",
          "implementations": [],
          "is_contract": false,
          "is_scam": false,
          "is_verified": false,
          "metadata": null,
          "name": null,
          "private_tags": [],
          "proxy_type": null,
          "public_tags": [],
          "reputation": "ok",
          "watchlist_names": []
        },
        "block_hash": "0x03a8690bf2c162ead5bd559a89d30f17b82f06d23a4937658381bfaa04439d3e",
        "block_number": 187369,
        "data": "0x0000000000000000000000000000000000000000000000000dbd2fc137a300000000000000000000000000000000000000000000033b19465ad2ffe5498a0a08",
        "decoded": {
          "method_call": "Sync(uint256 reserve0, uint256 reserve1)",
          "method_id": "cf2aa508",
          "parameters": [
            {
              "indexed": false,
              "name": "reserve0",
              "type": "uint256",
              "value": "990000000000000000"
            },
            {
              "indexed": false,
              "name": "reserve1",
              "type": "uint256",
              "value": "999901009800029797050092040"
            }
          ]
        },
        "index": 5,
        "smart_contract": {
          "ens_domain_name": null,
          "hash": "0x534DfB04a1A15924daB357694647e4f957543e8F",
          "implementations": [],
          "is_contract": false,
          "is_scam": false,
          "is_verified": false,
          "metadata": null,
          "name": null,
          "private_tags": [],
          "proxy_type": null,
          "public_tags": [],
          "reputation": "ok",
          "watchlist_names": []
        },
        "topics": [
          "0xcf2aa50876cdfbb541206f89af0ee78d44a2abf8d328e37fa4917f982149848a",
          null,
          null,
          null
        ],
        "transaction_hash": "0x52ce00852d100eac200066975fa6bd603e9b65803e59e0a117ce0360650914d3"
      },
      {
        "address": {
          "ens_domain_name": null,
          "hash": "0x07E8681aAE88CcEE4763D2bDC2A03F498C2e0d52",
          "implementations": [],
          "is_contract": false,
          "is_scam": false,
          "is_verified": false,
          "metadata": null,
          "name": null,
          "private_tags": [],
          "proxy_type": null,
          "public_tags": [],
          "reputation": "ok",
          "watchlist_names": []
        },
        "block_hash": "0x03a8690bf2c162ead5bd559a89d30f17b82f06d23a4937658381bfaa04439d3e",
        "block_number": 187369,
        "data": "0x00000000000000000000000000000000000000000000000000000918c47885da000000000000000000000000000000000000000000000000000009184e72a000",
        "decoded": {
          "method_call": "PriceUpdate(uint256 numerator, uint256 denominator)",
          "method_id": "92664190",
          "parameters": [
            {
              "indexed": false,
              "name": "numerator",
              "type": "uint256",
              "value": "10001980098010"
            },
            {
              "indexed": false,
              "name": "denominator",
              "type": "uint256",
              "value": "10000000000000"
            }
          ]
        },
        "index": 4,
        "smart_contract": {
          "ens_domain_name": null,
          "hash": "0x534DfB04a1A15924daB357694647e4f957543e8F",
          "implementations": [],
          "is_contract": false,
          "is_scam": false,
          "is_verified": false,
          "metadata": null,
          "name": null,
          "private_tags": [],
          "proxy_type": null,
          "public_tags": [],
          "reputation": "ok",
          "watchlist_names": []
        },
        "topics": [
          "0x92664190cca12aca9cd5309d87194bdda75bb51362d71c06e1a6f75c7c765711",
          null,
          null,
          null
        ],
        "transaction_hash": "0x52ce00852d100eac200066975fa6bd603e9b65803e59e0a117ce0360650914d3"
      },
      {
        "address": {
          "ens_domain_name": null,
          "hash": "0x07E8681aAE88CcEE4763D2bDC2A03F498C2e0d52",
          "implementations": [],
          "is_contract": false,
          "is_scam": false,
          "is_verified": false,
          "metadata": null,
          "name": null,
          "private_tags": [],
          "proxy_type": null,
          "public_tags": [],
          "reputation": "ok",
          "watchlist_names": []
        },
        "block_hash": "0x29a18c033af24b7d6184e45327d783e633f56ea359959d06307c9e428198cfba",
        "block_number": 187365,
        "data": "0x0000000000000000000000000000000000000000033b2e3c9fd0803ce8000000",
        "decoded": null,
        "index": 5,
        "smart_contract": {
          "ens_domain_name": null,
          "hash": "0xFB4E790C9f047c96a53eFf08b9F58E96E6730c6a",
          "implementations": [],
          "is_contract": false,
          "is_scam": false,
          "is_verified": false,
          "metadata": null,
          "name": null,
          "private_tags": [],
          "proxy_type": null,
          "public_tags": [],
          "reputation": "ok",
          "watchlist_names": []
        },
        "topics": [
          "0xc6ee03138ee70e66e2fbebe981db49ae4c4edb4f98b93150e0057e5d7c4fb193",
          null,
          null,
          null
        ],
        "transaction_hash": "0x9e82253ac59abf0dfa604ddeecdfc16fc571be5a11a1bcdd49674c218bde24ba"
      },
      {
        "address": {
          "ens_domain_name": null,
          "hash": "0x07E8681aAE88CcEE4763D2bDC2A03F498C2e0d52",
          "implementations": [],
          "is_contract": false,
          "is_scam": false,
          "is_verified": false,
          "metadata": null,
          "name": null,
          "private_tags": [],
          "proxy_type": null,
          "public_tags": [],
          "reputation": "ok",
          "watchlist_names": []
        },
        "block_hash": "0x29a18c033af24b7d6184e45327d783e633f56ea359959d06307c9e428198cfba",
        "block_number": 187365,
        "data": "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000033b2e3c9fd0803ce8000000",
        "decoded": {
          "method_call": "Sync(uint256 reserve0, uint256 reserve1)",
          "method_id": "cf2aa508",
          "parameters": [
            {
              "indexed": false,
              "name": "reserve0",
              "type": "uint256",
              "value": "0"
            },
            {
              "indexed": false,
              "name": "reserve1",
              "type": "uint256",
              "value": "1000000000000000000000000000"
            }
          ]
        },
        "index": 4,
        "smart_contract": {
          "ens_domain_name": null,
          "hash": "0xFB4E790C9f047c96a53eFf08b9F58E96E6730c6a",
          "implementations": [],
          "is_contract": false,
          "is_scam": false,
          "is_verified": false,
          "metadata": null,
          "name": null,
          "private_tags": [],
          "proxy_type": null,
          "public_tags": [],
          "reputation": "ok",
          "watchlist_names": []
        },
        "topics": [
          "0xcf2aa50876cdfbb541206f89af0ee78d44a2abf8d328e37fa4917f982149848a",
          null,
          null,
          null
        ],
        "transaction_hash": "0x9e82253ac59abf0dfa604ddeecdfc16fc571be5a11a1bcdd49674c218bde24ba"
      },
      {
        "address": {
          "ens_domain_name": null,
          "hash": "0x07E8681aAE88CcEE4763D2bDC2A03F498C2e0d52",
          "implementations": [],
          "is_contract": false,
          "is_scam": false,
          "is_verified": false,
          "metadata": null,
          "name": null,
          "private_tags": [],
          "proxy_type": null,
          "public_tags": [],
          "reputation": "ok",
          "watchlist_names": []
        },
        "block_hash": "0x29a18c033af24b7d6184e45327d783e633f56ea359959d06307c9e428198cfba",
        "block_number": 187365,
        "data": "0x000000000000000000000000000000000000000000000000000009184e72a000000000000000000000000000000000000000000000000000000009184e72a000",
        "decoded": {
          "method_call": "PriceUpdate(uint256 numerator, uint256 denominator)",
          "method_id": "92664190",
          "parameters": [
            {
              "indexed": false,
              "name": "numerator",
              "type": "uint256",
              "value": "10000000000000"
            },
            {
              "indexed": false,
              "name": "denominator",
              "type": "uint256",
              "value": "10000000000000"
            }
          ]
        },
        "index": 3,
        "smart_contract": {
          "ens_domain_name": null,
          "hash": "0xFB4E790C9f047c96a53eFf08b9F58E96E6730c6a",
          "implementations": [],
          "is_contract": false,
          "is_scam": false,
          "is_verified": false,
          "metadata": null,
          "name": null,
          "private_tags": [],
          "proxy_type": null,
          "public_tags": [],
          "reputation": "ok",
          "watchlist_names": []
        },
        "topics": [
          "0x92664190cca12aca9cd5309d87194bdda75bb51362d71c06e1a6f75c7c765711",
          null,
          null,
          null
        ],
        "transaction_hash": "0x9e82253ac59abf0dfa604ddeecdfc16fc571be5a11a1bcdd49674c218bde24ba"
      }
    ],
    "next_page_params": null
  }
  ````