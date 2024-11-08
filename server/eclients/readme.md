# Note
"eclients" means external clients, it contains all clients such as conflux-pay and so on.

# 用户付费类型
两种类型：
-  1. 预付费用户（实际是按照 api 支付 api 调用价格，sponsor 按照 cfx_price计费）
-  2. 后付费用户（实际是按照铸造个数收费（当前每个 mint 0.7 元）, 在数据库中 cfx_price 为0，这样就相当于代付免费）

然后两种用户都可以设置可欠费额度，但是当前只有后付费用户设置了可欠费额度。所以后付费相当于是 “后付费+按铸造个数收费” 的用户

3月底支持的按铸造个数收费

## 如何设置某用户为免费（财务不允许设置 cfx_price 为 0）
1. 设置 users.user_type 为 2
2. 设置 user_balances.cfx_price 为 0
3. 设置 user_api_quota.cost_type 3 值为无穷
