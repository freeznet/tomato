package controllers

// AudiencesController 处理 /push_audiences 接口的请求
type AudiencesController struct {
	ClassesController
}

// HandleFind 处理查找 push_audiences 请求
// @router / [get]
func (s *AudiencesController) HandleFind() {
	s.ClassName = "_Audience"
	s.ClassesController.HandleFind()
}

// HandleGet 处理获取指定 push_audiences 请求
// @router /:objectId [get]
func (s *AudiencesController) HandleGet() {
	s.ClassName = "_Audience"
	s.ObjectID = s.Ctx.Input.Param(":objectId")
	s.ClassesController.HandleGet()
}

// HandleCreate 处理 push_audiences 创建请求
// @router / [post]
func (s *AudiencesController) HandleCreate() {
	s.ClassName = "_Audience"
	s.ClassesController.HandleCreate()
}

// HandleUpdate 处理更新指定 push_audiences 请求
// @router /:objectId [put]
func (s *AudiencesController) HandleUpdate() {
	s.ClassName = "_Audience"
	s.ObjectID = s.Ctx.Input.Param(":objectId")
	s.ClassesController.HandleUpdate()
}

// HandleDelete 处理删除指定 session 请求
// @router /:objectId [delete]
func (s *AudiencesController) HandleDelete() {
	s.ClassName = "_Audience"
	s.ClassesController.HandleDelete()
}
