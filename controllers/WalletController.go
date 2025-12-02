package controllers

import (
	"github.com/gin-gonic/gin"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"uasd-wallet-manager/log"
	"uasd-wallet-manager/models"
)

const (
	PathStr = "m/44'/60'/0'/0/"
)

type WalletController struct {
}

func (con WalletController) GetWallet(c *gin.Context) {
	var baseResonse models.BaseResonseModel
	wallet, err := GetWalletInfo(c)
	if err != nil {
		baseResonse.Code = -1
		baseResonse.Message = err.Error()
		baseResonse.Data = nil
	} else {
		baseResonse.Code = 0
		baseResonse.Message = "success"
		baseResonse.Data = wallet
	}

	c.JSON(http.StatusOK, baseResonse)
}
func GetWalletInfo(c *gin.Context) (wallet models.Wallet, error error) {
	var walletInfo models.Wallet

	err := c.ShouldBindJSON(&walletInfo)
	if err != nil {
		log.Errorw("getWalletInfo Error:", zap.Error(err))
		return models.Wallet{}, err
	}
	hwallet, err := hdwalletInit()
	if err != nil {
		log.Errorw("init hdwallet  err", zap.Error(err))
		return models.Wallet{}, err
	}
	PathString := PathStr + strconv.Itoa(int(walletInfo.Path))
	path := hdwallet.MustParseDerivationPath(PathString)
	hdAccount, err := hwallet.Derive(path, false)
	if err != nil {
		log.Errorw("init hdwallet  err", zap.Error(err))
		return models.Wallet{}, err
	}
	walletInfo.Address = hdAccount.Address.String()
	return walletInfo, nil
}
func hdwalletInit() (*hdwallet.Wallet, error) {

	wallet, err := hdwallet.NewFromMnemonic(Key)
	if err != nil {
		log.Errorw("init hdwallet  err", zap.Error(err))
		//PostLark("hdwallet error")
		return nil, err
	}
	return wallet, nil
}
