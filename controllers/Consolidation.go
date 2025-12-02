package controllers

import (
	"bytes"
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
	"github.com/robfig/cron/v3"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"math/big"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"uasd-wallet-manager/contracts"
	"uasd-wallet-manager/log"
)

var tokLimit = decimal.NewFromInt(1e17).Mul(decimal.NewFromInt(1))
var usadLimit = decimal.NewFromInt(1e18).Mul(decimal.NewFromInt(1))
var j int = 100

type Call3 struct {
	Target       common.Address
	AllowFailure bool
	CallData     []byte
}

// 解析返回数据
var results []struct {
	Success    bool
	ReturnData []byte
}

func ConsolidationManagerStart() {
	log.Infow("consolidation manager start")
	cron := cron.New(cron.WithSeconds())
	cron.AddFunc("0 */2 * * * *", ConsolidationManager)
	cron.Start()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	cron.Stop()

}
func ConsolidationManager() {
	wallet, err := hdwalletInit()
	if err != nil {
		log.Errorw("init hdwallet  err", zap.Error(err))
		PostLark("hdwallet error")
		return
	}
	var addresses []common.Address
	for i := 0; i <= j; i++ {
		PathString := PathStr + strconv.Itoa(int(i))
		path := hdwallet.MustParseDerivationPath(PathString)
		hdAccount, err := wallet.Derive(path, false)
		if err != nil {
			return
		}
		addresses = append(addresses, hdAccount.Address)
	}
	Multicall(addresses)
}
func Multicall(addresses []common.Address) error {
	multicallA := common.HexToAddress(multicallAddr)
	multicallData, err := os.ReadFile("./contracts/abi/multicall3.abi")
	if err != nil {
		log.Errorw("Failed to read multiCall.abi file", zap.Error(err))
	}
	multicallABI, err := abi.JSON(bytes.NewReader(multicallData))
	if err != nil {
		log.Errorf("Failed to parse Multicall3 ABI: %v", err)
	}
	erc20Data, err := os.ReadFile("./contracts/abi/ERC20.abi")
	if err != nil {
		log.Errorw("Failed to read multiCall.abi file", zap.Error(err))
	}
	erc20ABI, err := abi.JSON(bytes.NewReader(erc20Data))
	if err != nil {
		log.Errorf("Failed to parse Multicall3 ABI: %v", err)
	}
	var calls []Call3
	for _, addr := range addresses {

		callData, err := erc20ABI.Pack("balanceOf", addr)
		if err != nil {
			panic(fmt.Sprintf("编码地址 %s 失败：%v", addr, err))
		}
		// 添加到调用列表（目标合约：ERC20 代币地址，调用数据：编码后的 balanceOf 参数）
		calls = append(calls, Call3{
			Target:   common.HexToAddress(usadAddr),
			CallData: callData,
		})
	}

	input, err := multicallABI.Pack("aggregate3", calls)
	if err != nil {
		log.Errorf("Failed to parse Multicall3 ABI: %v", err)
	}

	// 准备调用 message
	msg := ethereum.CallMsg{
		To:   &multicallA,
		Data: input,
	}
	client, err := ethclient.Dial(network[9200])
	if err != nil {

		return fmt.Errorf("连接节点失败: %w", err)
	}
	// 执行调用
	ctx := context.Background()
	output, err := client.CallContract(ctx, msg, nil)
	if err != nil {
		log.Errorf("Failed to parse Multicall3 ABI: %v", err)
	}

	err = multicallABI.UnpackIntoInterface(&results, "aggregate3", output)
	if err != nil {
		log.Errorf("Failed to UnpackIntoInterface: %v", err)
	}
	var successAddresses []common.Address
	var balances []decimal.Decimal
	var path []int
	for i, res := range results {
		if res.Success && len(res.ReturnData) == 32 {
			balance := new(big.Int).SetBytes(res.ReturnData)
			//log.Infow("监测",
			//	zap.Int("path", i),
			//	zap.Any("hash", balance),
			//)
			if decimal.NewFromBigInt(balance, 0).Cmp(usadLimit) >= 0 {
				successAddresses = append(successAddresses, addresses[i])
				balances = append(balances, decimal.NewFromBigInt(balance, 0))
				path = append(path, i)
			}

		} else {
			//log.Info("Address %s query failed or returned empty\n", addresses[i].Hex())
		}
	}
	for i, _ := range successAddresses {
		TransferUsad(path[i], balances[i].BigInt())
	}
	return nil
}
func TransferUsad(path int, amount *big.Int) error {

	wallet, err := hdwalletInit()
	if err != nil {
		log.Errorw("init hdwallet  err", zap.Error(err))
		PostLark("hdwallet error")
		return err
	}
	//check tok
	err = checkTok(wallet, path)
	if err != nil {
		fmt.Println(err)
		return err
	}
	PathStringFrom := PathStr + strconv.Itoa(path)
	pathFrom := hdwallet.MustParseDerivationPath(PathStringFrom)
	hdAccountfrom, err := wallet.Derive(pathFrom, false)
	if err != nil {
		return err
	}
	privateKey, err := wallet.PrivateKeyHex(hdAccountfrom)
	if err != nil {
		fmt.Println(err)
	}
	client, err := ethclient.Dial(network[9200])
	if err != nil {

		return fmt.Errorf("连接节点失败: %w", err)
	}
	privateKeyECDSA, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		fmt.Println(err)
		return fmt.Errorf("解析私钥失败: %w", err)
	}

	// 创建签名交易器
	auth, err := bind.NewKeyedTransactorWithChainID(privateKeyECDSA, big.NewInt(9200))
	if err != nil {
		return fmt.Errorf("创建签名器失败: %w", err)
	}
	//
	token, err := contracts.NewToken(common.HexToAddress(usadAddr), client)
	if err != nil {
		return fmt.Errorf("加载ERC20合约失败: %w", err)
	}
	tx, err := token.Transfer(auth, common.HexToAddress(consolidationAddr), amount)
	if err != nil {
		return fmt.Errorf("发送approve失败: %w", err)
	}
	// 等待交易完成
	_, err = bind.WaitMined(context.Background(), client, tx)
	if err != nil {
		log.Errorw("等待交易失败: %v", err)
	}
	log.Infow("Transfer success", zap.Int("from", path), zap.Any("hash:", tx.Hash()))

	return nil
}

func checkTok(wallet *hdwallet.Wallet, path int) error {
	PathStringFrom := PathStr + strconv.Itoa(path)
	pathFrom := hdwallet.MustParseDerivationPath(PathStringFrom)
	hdAccountfrom, err := wallet.Derive(pathFrom, false)
	if err != nil {
		return err
	}
	client, err := ethclient.Dial(network[9200])
	if err != nil {

		return fmt.Errorf("连接节点失败: %w", err)
	}
	latestBlockNum, err := client.BlockNumber(context.Background())
	if err != nil {
		return fmt.Errorf("获取最新区块号失败: %w", err)
	}
	balanceWei, err := client.BalanceAt(context.Background(), hdAccountfrom.Address, big.NewInt(int64(latestBlockNum)))
	if err != nil {
		panic(fmt.Sprintf("查询余额失败：%v", err))
	}
	if decimal.NewFromBigInt(balanceWei, 0).Cmp(tokLimit) < 0 {
		amount := tokLimit.Mul(decimal.NewFromInt(10)).BigInt()
		PathString0 := PathStr + strconv.Itoa(1)
		path0 := hdwallet.MustParseDerivationPath(PathString0)
		hdAccount0, err := wallet.Derive(path0, false)
		if err != nil {
			return err
		}

		// 1. 获取私钥
		privateKeyHex, err := wallet.PrivateKeyHex(hdAccount0)
		if err != nil {
			return fmt.Errorf("获取私钥失败: %w", err)
		}

		privateKeyECDSA, err := crypto.HexToECDSA(privateKeyHex)
		if err != nil {
			return fmt.Errorf("解析私钥失败: %w", err)
		}

		// 2. 创建交易签名器
		chainID := big.NewInt(9200)

		// 3. 获取账户nonce
		fromAddress := crypto.PubkeyToAddress(privateKeyECDSA.PublicKey)
		nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
		if err != nil {
			return fmt.Errorf("获取nonce失败: %w", err)
		}

		// 4. 获取网络推荐Gas参数
		gasPrice, err := client.SuggestGasPrice(context.Background())
		if err != nil {
			return fmt.Errorf("获取GasPrice失败: %w", err)
		}
		toAddress := common.HexToAddress(consolidationAddr)
		// 5. 估算GasLimit
		gasLimit, err := client.EstimateGas(context.Background(), ethereum.CallMsg{
			From:     fromAddress,
			To:       &toAddress,
			Value:    amount,
			GasPrice: gasPrice,
		})
		if err != nil {
			// 如果估算失败，使用安全默认值
			gasLimit = 21000
		}

		// 添加10%的缓冲
		gasLimit = gasLimit * 110 / 100

		// 6. 构造交易
		tx := types.NewTransaction(
			nonce,
			hdAccountfrom.Address,
			amount,
			gasLimit,
			gasPrice,
			nil,
		)

		// 7. 签名交易
		signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKeyECDSA)
		if err != nil {
			return fmt.Errorf("签名交易失败: %w", err)
		}

		// 8. 发送交易
		err = client.SendTransaction(context.Background(), signedTx)
		if err != nil {
			return fmt.Errorf("发送交易失败: %w", err)
		}

		// 9. 等待交易确认
		receipt, err := bind.WaitMined(context.Background(), client, signedTx)
		if err != nil {
			return fmt.Errorf("等待交易确认失败: %w", err)
		}

		if receipt.Status == 0 {
			return fmt.Errorf("交易执行失败，交易哈希: %s", signedTx.Hash().Hex())
		}

		log.Infow("ETH转账成功",
			zap.Int("path", path),
			zap.String("hash", signedTx.Hash().Hex()),
			zap.String("amount", amount.String()),
		)

	}
	return nil
}
