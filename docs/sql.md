# 设置用户免费(过期，财务不允许)
```
update user_balances set cfx_price=0 where user_id=USER_ID;
update user_api_quota set count_rollover=10000000 where user_id=USER_ID and cost_type in (2,3,4);
```