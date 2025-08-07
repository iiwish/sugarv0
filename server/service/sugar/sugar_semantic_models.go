package sugar

import (
	"context"
	"errors"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/sugar"
	sugarReq "github.com/flipped-aurora/gin-vue-admin/server/model/sugar/request"
)

type SugarSemanticModelsService struct{}

// CreateSugarSemanticModels 创建Sugar指标语义表记录
func (s *SugarSemanticModelsService) CreateSugarSemanticModels(ctx context.Context, model *sugar.SugarSemanticModels) (err error) {
	err = global.GVA_DB.Create(model).Error
	return err
}

// DeleteSugarSemanticModels 删除Sugar指标语义表记录
func (s *SugarSemanticModelsService) DeleteSugarSemanticModels(ctx context.Context, id string, userId string) (err error) {
	var model sugar.SugarSemanticModels
	if err = global.GVA_DB.Where("id = ?", id).First(&model).Error; err != nil {
		return errors.New("记录不存在")
	}
	if *model.CreatedBy != userId {
		return errors.New("无权删除")
	}
	err = global.GVA_DB.Delete(&sugar.SugarSemanticModels{}, "id = ?", id).Error
	return err
}

// DeleteSugarSemanticModelsByIds 批量删除Sugar指标语义表记录
func (s *SugarSemanticModelsService) DeleteSugarSemanticModelsByIds(ctx context.Context, ids []string, userId string) (err error) {
	err = global.GVA_DB.Where("id IN ? AND created_by = ?", ids, userId).Delete(&[]sugar.SugarSemanticModels{}).Error
	return err
}

// UpdateSugarSemanticModels 更新Sugar指标语义表记录
func (s *SugarSemanticModelsService) UpdateSugarSemanticModels(ctx context.Context, model sugar.SugarSemanticModels, userId string) (err error) {
	var oldModel sugar.SugarSemanticModels
	if err = global.GVA_DB.Where("id = ?", model.Id).First(&oldModel).Error; err != nil {
		return errors.New("记录不存在")
	}
	if *oldModel.CreatedBy != userId {
		return errors.New("无权更新")
	}
	err = global.GVA_DB.Model(&sugar.SugarSemanticModels{}).Where("id = ?", model.Id).Updates(&model).Error
	return err
}

// GetSugarSemanticModels 根据id获取Sugar指标语义表记录
func (s *SugarSemanticModelsService) GetSugarSemanticModels(ctx context.Context, id string, userId string) (model sugar.SugarSemanticModels, err error) {
	if err = global.GVA_DB.Where("id = ?", id).First(&model).Error; err != nil {
		return model, errors.New("记录不存在")
	}
	// TODO: 增加团队成员权限校验
	if *model.CreatedBy != userId {
		// return model, errors.New("无权查看")
	}
	return model, nil
}

// GetSugarSemanticModelsListByUser 分页获取Sugar指标语义表记录
func (s *SugarSemanticModelsService) GetSugarSemanticModelsListByUser(ctx context.Context, info sugarReq.SugarSemanticModelsSearch, userId string) (list []sugar.SugarSemanticModels, total int64, err error) {
	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)

	var teamIds []string
	err = global.GVA_DB.Table("sugar_team_members").Where("user_id = ?", userId).Pluck("team_id", &teamIds).Error
	if err != nil {
		return nil, 0, err
	}
	if len(teamIds) == 0 {
		return []sugar.SugarSemanticModels{}, 0, nil
	}

	db := global.GVA_DB.Model(&sugar.SugarSemanticModels{}).Where("team_id IN ?", teamIds)
	var sugarSemanticModelss []sugar.SugarSemanticModels

	err = db.Count(&total).Error
	if err != nil {
		return
	}

	if limit != 0 {
		db = db.Limit(limit).Offset(offset)
	}

	err = db.Order("updated_at desc").Find(&sugarSemanticModelss).Error
	return sugarSemanticModelss, total, err
}
