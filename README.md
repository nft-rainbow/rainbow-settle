# Rainbow-Settle

Rainbow-Settle 用于 rainbow 的费用结算，根据 header 计算每次请求的费用并结算。整体部署架构图见[这里](https://github.com/nft-rainbow/rainbow-apisix/tree/main/doc)

## Sever

结算服务

## Proxy

代理服务，转发所有 rainbow-apisix 传入的请求，根据 header 转发到不同的服务，下游服务包括  Rainbow-api, confura, scan 等。