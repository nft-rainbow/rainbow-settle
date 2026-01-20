# 设置用户免费(过期，财务不允许)
```sql
update user_balances set cfx_price=0 where user_id=USER_ID;
update user_api_quota set count_rollover=10000000 where user_id=USER_ID and cost_type in (2,3,4);
```

# 查某用户最后一次 mint
```sql
select a.user_id,m.created_at,m.contract from mint_tasks m left join contracts c on m.contract=c.address left join applications a on a.id=c.app_id where a.user_id=289 order by m.id desc limit 1;
```