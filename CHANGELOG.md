## ChangeLog

#### v1.0.8 (2026-09-06)

- deprecated Filter
- deprecated func NewFilter()
- add struct selectFilter instead Filter
- add func NewSelectFilter() instead NewFilter()
- deprecated method QueryBuilder.AndSelect()
- add method QueryBuilder.AddSelect()
- add method QueryBuilder.SelectFilter()
- add method QueryBuilder.AddSelectFilter()
- deprecated method QueryBuilder.FilterSelect()
- deprecated method QueryBuilder.AndFilterSelect()
- add method QueryBuilder.CrossJoin()


#### v1.0.7 (2026-08-17)

- add escape to method argsMap.ValueLike()
- add method QueryBuilder.Having()
- add method QueryBuilder.AndHaving()
- add method QueryBuilder.FilterHaving()
- add method QueryBuilder.AndFilterHaving()
- add method QueryBuilder.GetHaving()
- add method Filter.AddAliasForParam()
- add method Filter.AddAliasForParamWithJoin()


#### v1.0.6 (2026-06-01)

- add method argsMap.ValueTime()


#### v1.0.5 (2026-05-30)

- add method argsMap.ValueLike()


#### v1.0.4 (2026-05-26)

- add method QueryBuilder.AndFilterSelect()


#### v1.0.3 (2026-05-25)

- fix
- add method argsMap.Values()
- fix func NewArgs


#### v1.0.0 (2026-05-24)

First release