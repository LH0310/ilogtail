SQL 数据处理
---

该插件可以处理单层的结构化日志，使用selection clause筛选列，where clause筛选行，同时支持as重命名列，以及标量函数对列进行处理。

#### 参数说明

插件类型（type）为 `processor_sql`。

|参数|类型|必选或可选|参数说明|
|----|----|----|----|
|NoKeyError|bool|可选|无匹配的key是否记录，默认false。|
|SQL|string|必选|处理日志的 SQL 语句|

#### 示例

- 输入

```json
"timestamp":  "1234567890",
"nanosecond": "123456789",
"event_type": "js_error",
"idfa":       "abcdefg",
"user_agent": "Chrome on iOS. Mozilla/5.0 (iPhone; CPU iPhone OS 16_5_1 like Mac OS X)",
"action":     "click",
"element":    "#Button",
```

- 配置详情

```yaml
processors:
  - type: processor_sql
    SQL: - |
    SELECT 
	    CONCAT_WS(".", timestamp, nanosecond) AS event_time,
	    event_type,
	    MD5(idfa) AS idfa,
	    CASE 
	    	WHEN user_agent LIKE "%iPhone OS%" THEN "ios"
	    	ELSE "android"
	    END AS os,
	    action,
	    LOWER(element) AS element
    FROM 
	    log
    WHERE 
	    event_type = "js_error";
```

- 配置后结果

```json
"event_time": "1234567890.123456789",
"event_type": "js_error",
"idfa":       "7ac66c0f148de9519b8bd264312c4d64",
"os":         "ios",
"action":     "click",
"element":    "#button",
```
