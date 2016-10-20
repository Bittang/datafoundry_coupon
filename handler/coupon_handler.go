package handler

import (
	"github.com/asiainfoLDP/datafoundry_coupon/api"
	"github.com/asiainfoLDP/datafoundry_coupon/common"
	"github.com/asiainfoLDP/datafoundry_coupon/log"
	"github.com/asiainfoLDP/datafoundry_coupon/models"
	"github.com/julienschmidt/httprouter"
	"math/rand"
	"net/http"
	"time"
)

const letterBytes = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var logger = log.GetLogger()

//func init() {
//	mathrand.Seed(time.Now().UnixNano())
//}

func CreateCoupon(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	logger.Info("Request url: POST %v.", r.URL)
	logger.Info("Begin create coupon handler.")

	db := models.GetDB()
	if db == nil {
		logger.Warn("Get db is nil.")
		api.JsonResult(w, http.StatusInternalServerError, api.GetError(api.ErrorCodeDbNotInitlized), nil)
		return
	}

	coupon := &models.Coupon{}
	err := common.ParseRequestJsonInto(r, coupon)
	if err != nil {
		logger.Error("Parse body err: %v", err)
		api.JsonResult(w, http.StatusBadRequest, api.GetError2(api.ErrorCodeParseJsonFailed, err.Error()), nil)
		return
	}

	coupon.Serial = genUUID()
	coupon.Code = genUUID()

	logger.Debug("coupon: %v", coupon)

	//create coupon in database
	result, err := models.CreateCoupon(db, coupon)
	if err != nil {
		logger.Error("Create plan err: %v", err)
		api.JsonResult(w, http.StatusBadRequest, api.GetError2(api.ErrorCodeCreateCoupon, err.Error()), nil)
		return
	}

	logger.Info("End create coupon handler.")
	api.JsonResult(w, http.StatusOK, nil, result)
}

func DeleteCoupon(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	logger.Info("Request url: DELETE %v.", r.URL)

	logger.Info("Begin delete coupon handler.")

	db := models.GetDB()
	if db == nil {
		logger.Warn("Get db is nil.")
		api.JsonResult(w, http.StatusInternalServerError, api.GetError(api.ErrorCodeDbNotInitlized), nil)
		return
	}

	couponId := params.ByName("id")
	logger.Debug("Coupon id: %s.", couponId)

	// /delete in database
	err := models.DeleteCoupon(db, couponId)
	if err != nil {
		logger.Error("Delete coupon err: %v", err)
		api.JsonResult(w, http.StatusBadRequest, api.GetError2(api.ErrorCodeDeleteCoupon, err.Error()), nil)
		return
	}

	logger.Info("End delete coupon handler.")
	api.JsonResult(w, http.StatusOK, nil, nil)
}

//func ModifyPlan(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
//	logger.Info("Request url: PUT %v.", r.URL)
//
//	logger.Info("Begin modify plan handler.")
//
//	db := models.GetDB()
//	if db == nil {
//		logger.Warn("Get db is nil.")
//		api.JsonResult(w, http.StatusInternalServerError, api.GetError(api.ErrorCodeDbNotInitlized), nil)
//		return
//	}
//
//	plan := &models.Plan{}
//	err := common.ParseRequestJsonInto(r, plan)
//	if err != nil {
//		logger.Error("Parse body err: %v", err)
//		api.JsonResult(w, http.StatusBadRequest, api.GetError2(api.ErrorCodeParseJsonFailed, err.Error()), nil)
//		return
//	}
//	logger.Debug("Plan: %v", plan)
//
//	planId := params.ByName("id")
//	logger.Debug("Plan id: %s.", planId)
//
//	plan.Plan_id = planId
//
//	//update in database
//	err = models.ModifyPlan(db, plan)
//	if err != nil {
//		logger.Error("Modify plan err: %v", err)
//		api.JsonResult(w, http.StatusBadRequest, api.GetError2(api.ErrorCodeModifyPlan, err.Error()), nil)
//		return
//	}
//
//	logger.Info("End modify plan handler.")
//	api.JsonResult(w, http.StatusOK, nil, nil)
//}
//
func RetrieveCoupon(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	logger.Info("Request url: GET %v.", r.URL)

	logger.Info("Begin retrieve coupon handler.")

	db := models.GetDB()
	if db == nil {
		logger.Warn("Get db is nil.")
		api.JsonResult(w, http.StatusInternalServerError, api.GetError(api.ErrorCodeDbNotInitlized), nil)
		return
	}

	couponId := params.ByName("id")
	coupon, err := models.RetrieveCouponByID(db, couponId)
	if err != nil {
		logger.Error("Get coupon err: %v", err)
		api.JsonResult(w, http.StatusInternalServerError, api.GetError(api.ErrorCodeGetCoupon), nil)
		return
	}

	logger.Info("End retrieve coupon handler.")
	api.JsonResult(w, http.StatusOK, nil, coupon)
}

func QueryCouponList(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	logger.Info("Request url: GET %v.", r.URL)

	logger.Info("Begin retrieve coupon list handler.")

	db := models.GetDB()
	if db == nil {
		logger.Warn("Get db is nil.")
		api.JsonResult(w, http.StatusInternalServerError, api.GetError(api.ErrorCodeDbNotInitlized), nil)
		return
	}

	r.ParseForm()

	kind := r.Form.Get("kind")

	offset, size := api.OptionalOffsetAndSize(r, 30, 1, 100)
	orderBy := models.ValidateOrderBy(r.Form.Get("orderby"))
	sortOrder := models.ValidateSortOrder(r.Form.Get("sortorder"), false)

	count, apps, err := models.QueryCoupons(db, kind, orderBy, sortOrder, offset, size)
	if err != nil {
		api.JsonResult(w, http.StatusBadRequest, api.GetError2(api.ErrorCodeQueryCoupons, err.Error()), nil)
		return
	}

	logger.Info("End retrieve coupon list handler.")
	api.JsonResult(w, http.StatusOK, nil, api.NewQueryListResult(count, apps))
}

func UseCoupon(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	logger.Info("Request url: PUT %v.", r.URL)
	logger.Info("Begin use a coupon handler.")

	db := models.GetDB()
	if db == nil {
		logger.Warn("Get db is nil.")
		api.JsonResult(w, http.StatusInternalServerError, api.GetError(api.ErrorCodeDbNotInitlized), nil)
		return
	}

	serial := params.ByName("serial")
	code := params.ByName("code")

	rechargeInfo := &models.UseInfo{}
	err := common.ParseRequestJsonInto(r, rechargeInfo)
	if err != nil {
		logger.Error("Parse body err: %v", err)
		api.JsonResult(w, http.StatusBadRequest, api.GetError2(api.ErrorCodeParseJsonFailed, err.Error()), nil)
		return
	}
	rechargeInfo.Serial = serial
	rechargeInfo.Code = code
	rechargeInfo.Use_time = time.Now()

	result, err := models.UseCoupon(db, rechargeInfo)
	if err != nil {
		api.JsonResult(w, http.StatusBadRequest, api.GetError2(api.ErrorCodeUseCoupon, err.Error()), nil)
		return
	}

	logger.Info("End use a coupon handler.")
	api.JsonResult(w, http.StatusOK, nil, result)
}

func genUUID() string {
	b := make([]byte, 10)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

//func validateAppProvider(provider string, musBeNotBlank bool) (string, *Error) {
//	if musBeNotBlank || provider != "" {
//		// most 20 Chinese chars
//		provider_param, e := _mustStringParam("provider", provider, 60, StringParamType_General)
//		if e != nil {
//			return "", e
//		}
//		provider = provider_param
//	}
//
//	return provider, nil
//}
//
//func validateAppCategory(category string, musBeNotBlank bool) (string, *api.Error) {
//	if musBeNotBlank || category != "" {
//		// most 10 Chinese chars
//		category_param, e := _mustStringParam("category", category, 32, StringParamType_General)
//		if e != nil {
//			return "", e
//		}
//		category = category_param
//	}
//
//	return category, nil
//}
