package service

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/geoMonitor/model"
	geoReq "github.com/flipped-aurora/gin-vue-admin/server/plugin/geoMonitor/model/request"
	"golang.org/x/sync/errgroup"
)

type collector struct{}

type RunCollectionResponse struct {
	TaskID  uint                     `json:"taskId"`
	Status  string                   `json:"status"`
	Results []model.CollectionResult `json:"results"`
}

func (s *collector) Run(req geoReq.RunCollectionRequest, userID uint, userName string, authorityID uint) (RunCollectionResponse, error) {
	topic, err := Service.Topic.GetTopic(req.TopicID, userID, authorityID)
	if err != nil {
		return RunCollectionResponse{}, err
	}

	var platforms []model.Platform
	if err := global.GVA_DB.Where("id IN ? AND mode = ? AND status = ?", req.PlatformIDs, req.Mode, 1).Order("sort asc, id asc").Find(&platforms).Error; err != nil {
		return RunCollectionResponse{}, err
	}
	if len(platforms) == 0 {
		return RunCollectionResponse{}, fmt.Errorf("未找到可执行平台")
	}

	platformIDsJSON, _ := json.Marshal(req.PlatformIDs)
	startedAt := time.Now()
	task := model.CollectionTask{
		TopicID:         topic.ID,
		TopicName:       topic.Name,
		Prompt:          topic.Prompt,
		Mode:            req.Mode,
		Status:          TaskStatusRunning,
		PlatformIDs:     string(platformIDsJSON),
		RequestedBy:     userID,
		RequestedByName: userName,
		StartedAt:       &startedAt,
	}
	if err := global.GVA_DB.Create(&task).Error; err != nil {
		return RunCollectionResponse{}, err
	}

	results := make([]model.CollectionResult, 0, len(platforms))
	var mu sync.Mutex
	g := new(errgroup.Group)
	g.SetLimit(3)

	for _, platform := range platforms {
		platform := platform
		g.Go(func() error {
			output, collectErr := s.collectOne(task.ID, platform, topic.Prompt, req.Mode)
			record := model.CollectionResult{
				TaskID:         task.ID,
				PlatformID:     platform.ID,
				PlatformName:   platform.Name,
				PlatformCode:   platform.Code,
				Mode:           req.Mode,
				Prompt:         topic.Prompt,
				Answer:         output.Answer,
				Status:         ResultStatusSuccess,
				ErrorMsg:       output.ErrorMsg,
				Citations:      output.Citations,
				ScreenshotPath: output.ScreenshotPath,
				DurationMs:     output.DurationMs,
				RawResponse:    output.RawResponse,
				RunLog:         output.RunLog,
			}
			if collectErr != nil {
				record.Status = ResultStatusFailed
				record.ErrorMsg = collectErr.Error()
				if record.RunLog == "" {
					runLog := NewRunLog()
					runLog.Add("collect", "failed", collectErr.Error(), 0)
					record.RunLog = runLog.JSON()
				}
			}
			if err := global.GVA_DB.Create(&record).Error; err != nil {
				return err
			}
			mu.Lock()
			results = append(results, record)
			mu.Unlock()
			return nil
		})
	}

	groupErr := g.Wait()
	finishedAt := time.Now()
	finalStatus := summarizeTaskStatus(results)
	update := map[string]any{
		"status":      finalStatus,
		"finished_at": &finishedAt,
	}
	if groupErr != nil {
		update["error_msg"] = groupErr.Error()
		if finalStatus == TaskStatusDone {
			update["status"] = TaskStatusPartial
			finalStatus = TaskStatusPartial
		}
	}
	if err := global.GVA_DB.Model(&model.CollectionTask{}).Where("id = ?", task.ID).Updates(update).Error; err != nil {
		return RunCollectionResponse{}, err
	}

	return RunCollectionResponse{TaskID: task.ID, Status: finalStatus, Results: results}, nil
}

func (s *collector) collectOne(taskID uint, platform model.Platform, prompt string, mode string) (CollectOutput, error) {
	switch mode {
	case CollectModeAPI:
		return s.collectWithAPI(platform, prompt)
	case CollectModePlaywright:
		return s.collectWithPlaywright(platform, prompt, taskID)
	default:
		return CollectOutput{}, fmt.Errorf("不支持的采集模式: %s", mode)
	}
}

func (s *collector) GetTask(id uint, userID uint, authorityID uint) (model.CollectionTask, error) {
	var task model.CollectionTask
	db := global.GVA_DB.Where("id = ?", id)
	if authorityID != superAdminAuthorityID {
		db = db.Where("requested_by = ?", userID)
	}
	err := db.First(&task).Error
	return task, err
}

func (s *collector) GetResultList(taskID uint, page int, pageSize int, userID uint, authorityID uint) ([]model.CollectionResult, int64, error) {
	if _, err := s.GetTask(taskID, userID, authorityID); err != nil {
		return nil, 0, err
	}
	var list []model.CollectionResult
	var total int64
	db := global.GVA_DB.Model(&model.CollectionResult{}).Where("task_id = ?", taskID)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Order("id desc").Limit(pageSize).Offset(pageSize * (page - 1)).Find(&list).Error
	return list, total, err
}

func summarizeTaskStatus(results []model.CollectionResult) string {
	if len(results) == 0 {
		return TaskStatusFailed
	}
	successCount := 0
	for _, result := range results {
		if result.Status == ResultStatusSuccess {
			successCount++
		}
	}
	switch {
	case successCount == len(results):
		return TaskStatusDone
	case successCount == 0:
		return TaskStatusFailed
	default:
		return TaskStatusPartial
	}
}
