#######################################################################################################
# Set your parameters
#######################################################################################################

# # local
# apisix_addr=http://127.0.0.1:9180
# servers_domain=nftrainbow.me
# rainbow_api_addr=http://172.16.100.252:8080
# settle_addr=http://172.16.100.252:8091
# proxy_addr="172.16.100.252:8020"
# jwt_auth_env=local

# dev
apisix_addr=http://dev-apisix-admin.nftrainbow.cn
servers_domain=nftrainbow.cn
rainbow_api_addr=http://127.0.0.1:8080
settle_addr=http://127.0.0.1:8091
proxy_addr="172.18.0.1:8020"
jwt_auth_env=dev
apikey_confura_main="0rW8CEuqNvDaWNybiukVXK5kJp9GP3rdptimpqxu9bdc"
apikey_confura_test="0djrpfkthikrMfSkRzHDdAVD6biYJ42GaWopMkew3t6"

echo "开始配置apisix路由"


#######################################################################################################

# ******************** rainbow 使用的upstream *******************

# 添加upstream
curl $apisix_addr/apisix/admin/upstreams/100  \
-H 'X-API-KEY: edd1c9f034335f136f87ad84b625c8f1' -i -X PUT -d '
{
    "type":"roundrobin",
    "nodes":{
        "'${proxy_addr}'": 1
    }
}'

# 查upstream
curl $apisix_addr/apisix/admin/upstreams/100 -H 'X-API-KEY: edd1c9f034335f136f87ad84b625c8f1'


# ******************** rainbow 使用的路由 *******************

# rainbow open api
curl $apisix_addr/apisix/admin/routes/1000 -H 'X-API-KEY: edd1c9f034335f136f87ad84b625c8f1' -X PUT -d '
{
  "name": "rainbow-openapi",
  "desc": "rainbow open api 路由，只匹配openapi需要收费的api",
  "uri": "/*",
  "vars": [
    ["uri", "~~", "^/v1/(accounts|mints|transfers|burns|contracts|metadata|files|nft|tx)/.*$"]
  ],
  "host": "devapi.'${servers_domain}'",
  "plugins": {
    "ext-plugin-pre-req": {
       "conf": [
         {"name":"jwt-auth", "value":"{\"token_lookup\":\"header: Authorization\",\"app\":\"rainbow-api\",\"env\":\"'${jwt_auth_env}'\"}"},
         {"name":"rainbow-api-parser", "value":"{}"},
         {"name":"count", "value":"{}"},
         {"name":"rate-limit", "value":"{\"mode\":\"request\"}"}
       ]
    },
    "proxy-rewrite": {
      "headers": {
        "target_addr": "'${rainbow_api_addr}'"
      }
    },
    "ext-plugin-post-resp": {
       "conf": [
         {"name":"count","value":"{}"}
       ]
    } 
  },
  "upstream_id": "100",
  "priority": 400
}'

# exit 0

# TODO: rainbow api dashboard 收费相关接口
curl $apisix_addr/apisix/admin/routes/1100 -H 'X-API-KEY: edd1c9f034335f136f87ad84b625c8f1' -X PUT -d '
{
  "name": "rainbow-dashboard-api",
  "desc": "rainbow dashboard api 路由,只匹配dashboard需要收费的api",
  "uri": "/*",
  "vars": [
    ["uri", "~~", "^/dashboard/apps/*/(contracts|nft).*$"]
  ],
  "host": "dev.'${servers_domain}'",
  "methods": ["POST"],
  "plugins": {
    "ext-plugin-pre-req": {
       "conf": [
         {"name":"jwt-auth", "value":"{\"token_lookup\":\"header: Authorization\",\"app\":\"rainbow-api\",\"env\":\"local\"}"},
         {"name":"rainbow-api-parser", "value":"{}"},
         {"name":"count", "value":"{}"}
       ]
    },
    "proxy-rewrite": {
      "headers": {
        "target_addr": "'${rainbow_api_addr}'"
      }
    },
    "ext-plugin-post-resp": {
       "conf": [
         {"name":"count","value":"{}"}
       ]
    } 
  },
  "upstream_id": "100",
  "priority": 400
}'

# settle 服务
curl $apisix_addr/apisix/admin/routes/1150 -H 'X-API-KEY: edd1c9f034335f136f87ad84b625c8f1' -X PUT -d '
{
  "name": "rainbow-settle",
  "desc": "rainbow settle",
  "uri": "/*",
  "vars": [
    ["uri", "~~", "^/settle/.*$"]
  ],
  "host": "dev.'${servers_domain}'",
  "plugins": {
    "proxy-rewrite": {
      "headers": {
        "target_addr": "'${settle_addr}'"
      }
    }
  },
  "upstream_id": "100",
  "priority": 400
}'

# rainbow api 其它所有接口，包括 v1其它,swagger,debug,dashboard,settle,admin
curl $apisix_addr/apisix/admin/routes/1200 -H 'X-API-KEY: edd1c9f034335f136f87ad84b625c8f1' -X PUT -d '
{
  "name": "rainbow-api-normal",
  "desc": "rainbow api 基础路由，优先级最低，用于免费接口",
  "uri": "/*",
  "hosts": ["devapi.'${servers_domain}'","dev.'${servers_domain}'","devadmin.'${servers_domain}'"],
  "plugins": {
    "proxy-rewrite": {
      "headers": {
        "target_addr": "'${rainbow_api_addr}'"
      }
    }
  },
  "upstream_id": "100",
  "priority": 0
}'

# ******************** confura 路由 ********************
# cspace-main
curl $apisix_addr/apisix/admin/routes/2000 -H 'X-API-KEY: edd1c9f034335f136f87ad84b625c8f1' -X PUT -d '
{
  "name": "rpc-cspace-main",
  "desc": "confura core space main net",
  "uri": "/*",
  "host": "dev-rpc-cspace-main.'${servers_domain}'",
  "plugins": {
    "ext-plugin-pre-req": {
       "conf": [
         {"name":"apikey-auth", "value":"{\"lookup\":\"path\"}"},
         {"name":"confura-parser", "value":"{\"is_mainnet\":true,\"is_cspace\":true}"},
         {"name":"count", "value":"{}"},
         {"name":"rate-limit", "value":"{\"mode\":\"cost_type\"}"}
       ]
    },
    "proxy-rewrite": {
      "headers": {
        "target_url": "https://main.confluxrpc.com/'${apikey_confura_main}'"
      }
    },
    "ext-plugin-post-resp": {
       "conf": [
         {"name":"rpc-resp-format","value":"{}"}
       ]
    } 
  },
  "upstream_id": "100",
  "priority": 400
}'

# cspace-test
curl $apisix_addr/apisix/admin/routes/2100 -H 'X-API-KEY: edd1c9f034335f136f87ad84b625c8f1' -X PUT -d '
{
  "name": "rpc-cspace-test",
  "desc": "confura core space test net",
  "uri": "/*",
  "host": "dev-rpc-cspace-test.'${servers_domain}'",
  "plugins": {
    "ext-plugin-pre-req": {
       "conf": [
         {"name":"apikey-auth", "value":"{\"lookup\":\"path\"}"},
         {"name":"confura-parser", "value":"{\"is_mainnet\":false,\"is_cspace\":true}"},
         {"name":"count", "value":"{}"},
         {"name":"rate-limit", "value":"{\"mode\":\"cost_type\"}"}
       ]
    },
    "proxy-rewrite": {
      "headers": {
        "target_url": "https://test.confluxrpc.com/'${apikey_confura_test}'"
      }
    },
    "ext-plugin-post-resp": {
       "conf": [
         {"name":"rpc-resp-format","value":"{}"}
       ]
    } 
  },
  "upstream_id": "100",
  "priority": 400
}'

# espace-main
curl $apisix_addr/apisix/admin/routes/2200 -H 'X-API-KEY: edd1c9f034335f136f87ad84b625c8f1' -X PUT -d '
{
  "name": "rpc-espace-main",
  "desc": "confura espace mainnet",
  "uri": "/*",
  "host": "dev-rpc-espace-main.'${servers_domain}'",
  "plugins": {
    "ext-plugin-pre-req": {
       "conf": [
         {"name":"apikey-auth", "value":"{\"lookup\":\"path\"}"},
         {"name":"confura-parser", "value":"{\"is_mainnet\":true,\"is_cspace\":false}"},
         {"name":"count", "value":"{}"},
         {"name":"rate-limit", "value":"{\"mode\":\"cost_type\"}"}
       ]
    },
    "proxy-rewrite": {
      "headers": {
        "target_url": "https://evm.confluxrpc.com/'${apikey_confura_main}'"
      }
    },
    "ext-plugin-post-resp": {
       "conf": [
         {"name":"rpc-resp-format","value":"{}"}
       ]
    } 
  },
  "upstream_id": "100",
  "priority": 400
}'

# espace-test
curl $apisix_addr/apisix/admin/routes/2300 -H 'X-API-KEY: edd1c9f034335f136f87ad84b625c8f1' -X PUT -d '
{
  "name": "rpc-espace-test",
  "desc": "confura espace testnet",
  "uri": "/*",
  "host": "dev-rpc-espace-test.'${servers_domain}'",
  "plugins": {
    "ext-plugin-pre-req": {
       "conf": [
         {"name":"apikey-auth", "value":"{\"lookup\":\"path\"}"},
         {"name":"confura-parser", "value":"{\"is_mainnet\":false,\"is_cspace\":false}"},
         {"name":"count", "value":"{}"},
         {"name":"rate-limit", "value":"{\"mode\":\"cost_type\"}"}
       ]
    },
    "proxy-rewrite": {
      "headers": {
        "target_url": "https://evmtestnet.confluxrpc.com/'${apikey_confura_test}'"
      }
    },
    "ext-plugin-post-resp": {
       "conf": [
         {"name":"rpc-resp-format","value":"{}"}
       ]
    } 
  },
  "upstream_id": "100",
  "priority": 400
}'

echo "配置apisix路由完成"
# *************************** 证书相关 ***********************************

# ssh证书生成

# openssl req -new -out server.csr -key server.key -subj "/C=CN/ST=BeiJing/L=BeiJing/O=blockchain/OU=conflux/CN=api.rainbow.com


# # ***************************** DEV 环境 ********************************
# 1. 将 $servers_domain 修改为 nftrainbow.cn
# 2. 将 127.0.0.1:9180 修改为 dev-apisix-admin.nftrainbow.cn
# 3. 将 upstream 修改为 172.18.0.1:8020
# 4. rainbow-api request-rewrite的header 修改为 172.18.0.1.8080
# 5. plugins 增加 http logger
#     "http-logger": {
#       "_meta": {
#         "disable": false
#       },
#       "include_req_body": true,
#       "include_resp_body": true,
#       "uri": "http://172.18.0.1:19080/logs/rconsole"
#     }
