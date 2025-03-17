

我修改了什么

swag是使用的方法上面的注释来生成api文档的，在日常开发中，我们还要去维护方法上面的注释。

示例,如果我们请求参数有变化则要手动维护.

```
// AddCategory
// @Tags      AddCategory
// @Summary   添加媒体库分类
// @Security  AttachmentCategory
// @accept    application/json
// @Produce   application/json
// @Param     data  body      example.ExaAttachmentCategory  true  "媒体库分类数据"
// @Success   200   {object}  response.Response{msg=string}   "添加媒体库分类"
// @Router    /attachmentCategory/addCategory [post]
```



实际开发中我们可以根据框架获取参数的方法，通过ast语法树分析来根据代码来生成文档。比如使用gin





```
├─api
│  └─account
├─model
│  ├─request
│  └─response
├─router
└─service

```

