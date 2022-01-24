package rpc

import (
	"context"
	"math/bits"

	sdk "github.com/Conflux-Chain/go-conflux-sdk"
	"github.com/Conflux-Chain/go-conflux-sdk/rpc"
	"github.com/Conflux-Chain/go-conflux-sdk/types"
	cimetrics "github.com/conflux-chain/conflux-infura/metrics"
	"github.com/conflux-chain/conflux-infura/node"
	"github.com/conflux-chain/conflux-infura/relay"
	"github.com/conflux-chain/conflux-infura/store"
	"github.com/conflux-chain/conflux-infura/util"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/metrics"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	// Flyweight objects
	emptyEpochs         = []*types.Epoch{}
	emptyLogs           = []types.Log{}
	emptyDepositInfos   = []types.DepositInfo{}
	emptyVoteStakeInfos = []types.VoteStakeInfo{}
	emptyHashes         = []types.Hash{}
	emptyRewards        = []types.RewardInfo{}
	emptySponsorInfo    = types.SponsorInfo{}
	emptyBlock          = types.Block{}

	hitStatsCollector = cimetrics.NewHitStatsCollector()
)

type cfxAPI struct {
	provider         *node.ClientProvider
	inputEpochMetric cimetrics.InputEpochMetric
	handler          cfxHandler
	relayer          *relay.TxnRelayer
}

func newCfxAPI(
	provider *node.ClientProvider, handler cfxHandler, relayer *relay.TxnRelayer,
) *cfxAPI {
	return &cfxAPI{
		provider: provider,
		handler:  handler,
		relayer:  relayer,
	}
}

func toSlice(epoch *types.Epoch) []*types.Epoch {
	if epoch == nil {
		return emptyEpochs
	}

	return []*types.Epoch{epoch}
}

func (api *cfxAPI) GasPrice(ctx context.Context) (*hexutil.Big, error) {
	cfx, err := api.provider.GetClientByIP(ctx)
	if err != nil {
		return nil, err
	}

	return cfx.GetGasPrice()
}

func (api *cfxAPI) EpochNumber(ctx context.Context, epoch *types.Epoch) (*hexutil.Big, error) {
	cfx, err := api.provider.GetClientByIP(ctx)
	if err != nil {
		return nil, err
	}

	api.inputEpochMetric.Update(epoch, "cfx_epochNumber", cfx)
	return cfx.GetEpochNumber(toSlice(epoch)...)
}

func (api *cfxAPI) GetBalance(ctx context.Context, address types.Address, epoch *types.Epoch) (*hexutil.Big, error) {
	cfx, err := api.provider.GetClientByIP(ctx)
	if err != nil {
		return nil, err
	}

	api.inputEpochMetric.Update(epoch, "cfx_getBalance", cfx)
	return cfx.GetBalance(address, toSlice(epoch)...)
}

func (api *cfxAPI) GetAdmin(ctx context.Context, contract types.Address, epoch *types.Epoch) (*types.Address, error) {
	cfx, err := api.provider.GetClientByIP(ctx)
	if err != nil {
		return nil, err
	}

	api.inputEpochMetric.Update(epoch, "cfx_getAdmin", cfx)
	return cfx.GetAdmin(contract, toSlice(epoch)...)
}

func (api *cfxAPI) GetSponsorInfo(ctx context.Context, contract types.Address, epoch *types.Epoch) (types.SponsorInfo, error) {
	cfx, err := api.provider.GetClientByIP(ctx)
	if err != nil {
		return emptySponsorInfo, err
	}

	api.inputEpochMetric.Update(epoch, "cfx_getSponsorInfo", cfx)
	return cfx.GetSponsorInfo(contract, toSlice(epoch)...)
}

func (api *cfxAPI) GetStakingBalance(ctx context.Context, address types.Address, epoch *types.Epoch) (*hexutil.Big, error) {
	cfx, err := api.provider.GetClientByIP(ctx)
	if err != nil {
		return nil, err
	}

	api.inputEpochMetric.Update(epoch, "cfx_getStakingBalance", cfx)
	return cfx.GetStakingBalance(address, toSlice(epoch)...)
}

func (api *cfxAPI) GetDepositList(ctx context.Context, address types.Address, epoch *types.Epoch) ([]types.DepositInfo, error) {
	cfx, err := api.provider.GetClientByIP(ctx)
	if err != nil {
		return emptyDepositInfos, err
	}

	api.inputEpochMetric.Update(epoch, "cfx_getDepositList", cfx)
	return cfx.GetDepositList(address, toSlice(epoch)...)
}

func (api *cfxAPI) GetVoteList(ctx context.Context, address types.Address, epoch *types.Epoch) ([]types.VoteStakeInfo, error) {
	cfx, err := api.provider.GetClientByIP(ctx)
	if err != nil {
		return emptyVoteStakeInfos, err
	}

	api.inputEpochMetric.Update(epoch, "cfx_getVoteList", cfx)
	return cfx.GetVoteList(address, toSlice(epoch)...)
}

func (api *cfxAPI) GetCollateralForStorage(ctx context.Context, address types.Address, epoch *types.Epoch) (*hexutil.Big, error) {
	cfx, err := api.provider.GetClientByIP(ctx)
	if err != nil {
		return nil, err
	}

	api.inputEpochMetric.Update(epoch, "cfx_getCollateralForStorage", cfx)
	return cfx.GetCollateralForStorage(address, toSlice(epoch)...)
}

func (api *cfxAPI) GetCode(ctx context.Context, contract types.Address, epoch *types.Epoch) (hexutil.Bytes, error) {
	cfx, err := api.provider.GetClientByIP(ctx)
	if err != nil {
		return nil, err
	}

	api.inputEpochMetric.Update(epoch, "cfx_getCode", cfx)
	return cfx.GetCode(contract, toSlice(epoch)...)
}

func (api *cfxAPI) GetStorageAt(ctx context.Context, address types.Address, position types.Hash, epoch *types.Epoch) (hexutil.Bytes, error) {
	cfx, err := api.provider.GetClientByIP(ctx)
	if err != nil {
		return nil, err
	}

	api.inputEpochMetric.Update(epoch, "cfx_getStorageAt", cfx)
	return cfx.GetStorageAt(address, position, toSlice(epoch)...)
}

func (api *cfxAPI) GetStorageRoot(ctx context.Context, address types.Address, epoch *types.Epoch) (*types.StorageRoot, error) {
	cfx, err := api.provider.GetClientByIP(ctx)
	if err != nil {
		return nil, err
	}

	api.inputEpochMetric.Update(epoch, "cfx_getStorageRoot", cfx)
	return cfx.GetStorageRoot(address, toSlice(epoch)...)
}

func (api *cfxAPI) GetBlockByHash(ctx context.Context, blockHash types.Hash, includeTxs bool) (interface{}, error) {
	logger := logrus.WithFields(logrus.Fields{"blockHash": blockHash, "includeTxs": includeTxs})

	if !util.IsValidHashStr(blockHash.String()) { // directly fall through to fullnode delegation to conform rpc error
		goto delegation
	}

	if !util.IsInterfaceValNil(api.handler) {
		isStoreHit := false
		defer func(isHit *bool) {
			hitStatsCollector.CollectHitStats("infura/rpc/call/cfx_getBlockByHash/store/hitratio", *isHit)
		}(&isStoreHit)

		block, err := api.handler.GetBlockByHash(ctx, blockHash, includeTxs)
		if err == nil {
			logger.Debug("Loading epoch data for cfx_getBlockByHash hit in the store")

			isStoreHit = true
			return block, err
		}

		logger.WithError(err).Debug("Loading epoch data for cfx_getBlockByHash hit missed from the store")
	}

delegation:
	cfx, err := api.provider.GetClientByIP(ctx)
	if err != nil {
		logger.WithError(err).Debug("Failed to delegate cfx_getBlockByHash rpc request to fullnode")
		return nil, err
	}

	logger.WithField("fullnode", cfx.GetNodeURL()).Debug("Delegating cfx_getBlockByHash rpc request to fullnode")

	if includeTxs {
		metrics.GetOrRegisterGauge("rpc/cfx_getBlockByHash/details", nil).Inc(1)
		return cfx.GetBlockByHash(blockHash)
	}

	return cfx.GetBlockSummaryByHash(blockHash)
}

func (api *cfxAPI) GetBlockByHashWithPivotAssumption(ctx context.Context, blockHash, pivotHash types.Hash, epoch hexutil.Uint64) (types.Block, error) {
	cfx, err := api.provider.GetClientByIP(ctx)
	if err != nil {
		return emptyBlock, err
	}

	return cfx.GetBlockByHashWithPivotAssumption(blockHash, pivotHash, epoch)
}

func (api *cfxAPI) GetBlockByEpochNumber(ctx context.Context, epoch types.Epoch, includeTxs bool) (interface{}, error) {
	logger := logrus.WithFields(logrus.Fields{"epoch": epoch, "includeTxs": includeTxs})

	if !util.IsInterfaceValNil(api.handler) {
		isStoreHit := false
		defer func(isHit *bool) {
			hitStatsCollector.CollectHitStats("infura/rpc/call/cfx_getBlockByEpochNumber/store/hitratio", *isHit)
		}(&isStoreHit)

		block, err := api.handler.GetBlockByEpochNumber(ctx, &epoch, includeTxs)
		if err == nil {
			logger.Debug("Loading epoch data for cfx_getBlockByEpochNumber hit in the store")

			isStoreHit = true
			return block, err
		}

		logger.WithError(err).Debug("Loading epoch data for cfx_getBlockByEpochNumber hit missed from the store")
	}

	cfx, err := api.provider.GetClientByIP(ctx)
	if err != nil {
		logger.WithError(err).Debug("Failed to delegate cfx_getBlockByEpochNumber rpc request to fullnode")
		return nil, err
	}

	logger.WithField("fullnode", cfx.GetNodeURL()).Debug("Delegating cfx_getBlockByEpochNumber rpc request to fullnode")
	api.inputEpochMetric.Update(&epoch, "cfx_getBlockByEpochNumber", cfx)

	if includeTxs {
		metrics.GetOrRegisterGauge("rpc/cfx_getBlockByEpochNumber/details", nil).Inc(1)
		return cfx.GetBlockByEpoch(&epoch)
	}

	return cfx.GetBlockSummaryByEpoch(&epoch)
}

func (api *cfxAPI) GetBlockByBlockNumber(ctx context.Context, blockNumer hexutil.Uint64, includeTxs bool) (block interface{}, err error) {
	logger := logrus.WithFields(logrus.Fields{"blockNumber": blockNumer, "includeTxs": includeTxs})

	if !util.IsInterfaceValNil(api.handler) {
		isStoreHit := false
		defer func(isHit *bool) {
			hitStatsCollector.CollectHitStats("infura/rpc/call/cfx_getBlockByBlockNumber/store/hitratio", *isHit)
		}(&isStoreHit)

		block, err := api.handler.GetBlockByBlockNumber(ctx, blockNumer, includeTxs)
		if err == nil {
			logger.Debug("Loading epoch data for cfx_getBlockByBlockNumber hit in the store")

			isStoreHit = true
			return block, err
		}

		logger.WithError(err).Debug("Loading epoch data for cfx_getBlockByBlockNumber hit missed from the store")
	}

	cfx, err := api.provider.GetClientByIP(ctx)
	if err != nil {
		logger.WithError(err).Debug("Failed to delegate cfx_getBlockByBlockNumber rpc request to fullnode")
		return nil, err
	}

	if includeTxs {
		metrics.GetOrRegisterGauge("rpc/cfx_getBlockByBlockNumber/details", nil).Inc(1)
		return cfx.GetBlockByBlockNumber(blockNumer)
	}

	logger.WithField("fullnode", cfx.GetNodeURL()).Debug("Delegating cfx_getBlockByBlockNumber rpc request to fullnode")
	return cfx.GetBlockSummaryByBlockNumber(blockNumer)
}

func (api *cfxAPI) GetBestBlockHash(ctx context.Context) (types.Hash, error) {
	cfx, err := api.provider.GetClientByIP(ctx)
	if err != nil {
		return "", err
	}

	return cfx.GetBestBlockHash()
}

func (api *cfxAPI) GetNextNonce(ctx context.Context, address types.Address, epoch *types.Epoch) (*hexutil.Big, error) {
	cfx, err := api.provider.GetClientByIP(ctx)
	if err != nil {
		return nil, err
	}

	api.inputEpochMetric.Update(epoch, "cfx_getNextNonce", cfx)
	return cfx.GetNextNonce(address, toSlice(epoch)...)
}

func (api *cfxAPI) SendRawTransaction(ctx context.Context, signedTx hexutil.Bytes) (types.Hash, error) {
	cfx, err := api.provider.GetClientByIP(ctx)
	if err != nil {
		return "", err
	}

	txHash, err := cfx.SendRawTransaction(signedTx)
	if err == nil {
		// relay transaction broadcasting asynchronously
		if !api.relayer.AsyncRelay(signedTx) {
			logrus.Warn("Transaction relay pool is full")
		}
	}

	return txHash, err
}

func (api *cfxAPI) Call(ctx context.Context, request types.CallRequest, epoch *types.Epoch) (hexutil.Bytes, error) {
	cfx, err := api.provider.GetClientByIP(ctx)
	if err != nil {
		return nil, err
	}

	api.inputEpochMetric.Update(epoch, "cfx_call", cfx)
	return cfx.Call(request, epoch)
}

func (api *cfxAPI) GetLogs(ctx context.Context, filter types.LogFilter) ([]types.Log, error) {
	logger := logrus.WithField("filter", filter)

	cfx, err := api.provider.GetClientByIP(ctx)
	if err != nil {
		logger.WithError(err).Debug("Failed to get available cfx client for cfx_getLogs rpc request")
		return emptyLogs, err
	}

	if err := api.validateLogFilter(cfx, &filter); err != nil {
		logger.WithError(err).Debug("Invalid log filter parameter for cfx_getLogs rpc request")
		return emptyLogs, err
	}

	// TODO optimize cfx_getLogs metrics with asynchronization to minimize side effect for rpc request
	api.inputEpochMetric.Update(filter.FromEpoch, "cfx_getLogs/from", cfx)
	api.inputEpochMetric.Update(filter.ToEpoch, "cfx_getLogs/to", cfx)

	sfilter, ok := store.ParseLogFilter(cfx, &filter)
	if ok && !util.IsInterfaceValNil(api.handler) {
		isStoreHit := false
		defer func(isHit *bool) {
			hitStatsCollector.CollectHitStats("infura/rpc/call/cfx_getLogs/store/hitratio", *isHit)
		}(&isStoreHit)

		if logs, err := api.handler.GetLogs(ctx, sfilter); err == nil {
			// return empty slice rather than nil to comply with fullnode
			if logs == nil {
				logs = emptyLogs
			}

			logger.Debug("Loading epoch data for cfx_getLogs hit in the store")

			isStoreHit = true
			return logs, nil
		}

		logger.WithError(err).Debug("Loading epoch data for cfx_getLogs hit missed from the store")

		// Logs already pruned from database? If so, we'd rather not delegate this request to
		// the fullnode, as it might crush our fullnode.
		if errors.Is(err, store.ErrAlreadyPruned) {
			logger.Info("Epoch log data already pruned for cfx_getLogs to return")
			return emptyLogs, errors.New("failed to get stale epoch logs (data too old)")
		}
	}

	logger.WithField("fullnode", cfx.GetNodeURL()).Debug("Delegating cfx_getLogs rpc request to fullnode")

	// for any error, delegate request to full node, including:
	// 1. database level error
	// 2. record not found (log range mismatch)
	return cfx.GetLogs(filter)
}

func (api *cfxAPI) validateLogFilter(cfx sdk.ClientOperator, filter *types.LogFilter) error {
	var filterFlag store.LogFilterType // filter type set flag bitwise

	if filter.FromEpoch != nil || filter.ToEpoch != nil { // check if epoch range provided
		filterFlag |= store.LogFilterTypeEpochRange
	}

	if filter.FromBlock != nil || filter.ToBlock != nil { // check if block range provided
		filterFlag |= store.LogFilterTypeBlockRange
	}

	if len(filter.BlockHashes) != 0 { // check if block hashes provided
		filterFlag |= store.LogFilterTypeBlockHashes
	}

	if bits.OnesCount(uint(filterFlag)) > 1 { // different types of log filters are mutual exclusion
		return errInvalidLogFilter
	}

	switch {
	case filterFlag&store.LogFilterTypeBlockHashes != 0: // validate block hash log filter
		if len(filter.BlockHashes) > store.MaxLogBlockHashesSize {
			return errExceedLogFilterBlockHashLimit(len(filter.BlockHashes))
		}
	case filterFlag&store.LogFilterTypeBlockRange != 0: // validate block range log filter
		if filter.FromBlock == nil || filter.ToBlock == nil { // both fromBlock and toBlock must be provided
			return errInvalidLogFilter
		}

		fromBlock := filter.FromBlock.ToInt().Uint64()
		toBlock := filter.ToBlock.ToInt().Uint64()

		if fromBlock > toBlock {
			return errors.New("invalid block range (from block larger than to epoch)")
		}

		if count := toBlock - fromBlock + 1; count > store.MaxLogBlockRange {
			return errors.Errorf("block range exceeds maximum value %v", store.MaxLogBlockRange)
		}
	default: // validate epoch range log filter as default
		uniformLogFilterEpochRange(cfx, filter)

		epochFrom, ok1 := filter.FromEpoch.ToInt()
		epochTo, ok2 := filter.ToEpoch.ToInt()
		if ok1 && ok2 {
			epochFrom := epochFrom.Uint64()
			epochTo := epochTo.Uint64()

			if epochFrom > epochTo {
				return errors.New("invalid epoch range (from epoch larger than to epoch)")
			}

			if count := epochTo - epochFrom + 1; count > store.MaxLogEpochRange {
				return errors.Errorf("epoch range exceeds maximum value %v", store.MaxLogEpochRange)
			}
		}
	}

	if filter.Limit != nil && uint64(*filter.Limit) > store.MaxLogLimit {
		return errors.Errorf("limit field exceed the maximum value %v", store.MaxLogLimit)
	}

	return nil
}

// uniformLogFilterEpochRange sets default epoch range if not set and convert to numbered epoch if necessary
func uniformLogFilterEpochRange(cfx sdk.ClientOperator, filter *types.LogFilter) {
	if filter.FromEpoch == nil { // if no from epoch provided, set latest checkpoint epoch as default
		filter.FromEpoch = types.EpochLatestCheckpoint
	}

	if epoch, err := util.ConvertToNumberedEpoch(cfx, filter.FromEpoch); err == nil {
		filter.FromEpoch = epoch
	}

	if filter.ToEpoch == nil { // if no to epoch provided, set latest state epoch as default
		filter.ToEpoch = types.EpochLatestState
	}

	if epoch, err := util.ConvertToNumberedEpoch(cfx, filter.ToEpoch); err == nil {
		filter.ToEpoch = epoch
	}
}

func (api *cfxAPI) GetTransactionByHash(ctx context.Context, txHash types.Hash) (*types.Transaction, error) {
	logger := logrus.WithFields(logrus.Fields{"txHash": txHash})

	if !util.IsValidHashStr(txHash.String()) {
		goto delegation
	}

	if !util.IsInterfaceValNil(api.handler) {
		isStoreHit := false
		defer func(isHit *bool) {
			hitStatsCollector.CollectHitStats("infura/rpc/call/cfx_getTransactionByHash/store/hitratio", *isHit)
		}(&isStoreHit)

		tx, err := api.handler.GetTransactionByHash(ctx, txHash)
		if err == nil {
			logger.Debug("Loading epoch data for cfx_getTransactionByHash hit in the store")

			isStoreHit = true
			return tx, err
		}

		logger.WithError(err).Debug("Loading epoch data for cfx_getTransactionByHash hit missed from the store")
	}

delegation:
	cfx, err := api.provider.GetClientByIP(ctx)
	if err != nil {
		logger.WithError(err).Debug("Failed to delegate cfx_getTransactionByHash rpc request to fullnode")
		return nil, err
	}

	logger.WithField("fullnode", cfx.GetNodeURL()).Debug("Delegating cfx_getTransactionByHash rpc request to fullnode")
	return cfx.GetTransactionByHash(txHash)
}

func (api *cfxAPI) EstimateGasAndCollateral(ctx context.Context, request types.CallRequest, epoch *types.Epoch) (types.Estimate, error) {
	cfx, err := api.provider.GetClientByIP(ctx)
	if err != nil {
		return types.Estimate{}, err
	}

	api.inputEpochMetric.Update(epoch, "cfx_estimateGasAndCollateral", cfx)
	return cfx.EstimateGasAndCollateral(request, toSlice(epoch)...)
}

func (api *cfxAPI) CheckBalanceAgainstTransaction(ctx context.Context, account, contract types.Address, gas, price, storage *hexutil.Big, epoch *types.Epoch) (types.CheckBalanceAgainstTransactionResponse, error) {
	cfx, err := api.provider.GetClientByIP(ctx)
	if err != nil {
		return types.CheckBalanceAgainstTransactionResponse{}, err
	}

	api.inputEpochMetric.Update(epoch, "cfx_checkBalanceAgainstTransaction", cfx)
	return cfx.CheckBalanceAgainstTransaction(account, contract, gas, price, storage, toSlice(epoch)...)
}

func (api *cfxAPI) GetBlocksByEpoch(ctx context.Context, epoch types.Epoch) ([]types.Hash, error) {
	logger := logrus.WithFields(logrus.Fields{"epoch": epoch})

	if !util.IsInterfaceValNil(api.handler) {
		isStoreHit := false
		defer func(isHit *bool) {
			hitStatsCollector.CollectHitStats("infura/rpc/call/cfx_getBlocksByEpoch/store/hitratio", *isHit)
		}(&isStoreHit)

		blockHashes, err := api.handler.GetBlocksByEpoch(ctx, &epoch)
		if err == nil {
			logger.Debug("Loading epoch data for cfx_getBlocksByEpoch hit in the store")

			isStoreHit = true
			return blockHashes, err
		}

		logger.WithError(err).Debug("Loading epoch data for cfx_getBlocksByEpoch hit missed from the store")
	}

	cfx, err := api.provider.GetClientByIP(ctx)
	if err != nil {
		logger.WithError(err).Debug("Failed to delegate cfx_getBlocksByEpoch rpc request to fullnode")
		return emptyHashes, err
	}

	logger.WithField("fullnode", cfx.GetNodeURL()).Debug("Delegating cfx_getBlocksByEpoch rpc request to fullnode")
	api.inputEpochMetric.Update(&epoch, "cfx_getBlocksByEpoch", cfx)

	return cfx.GetBlocksByEpoch(&epoch)
}

func (api *cfxAPI) GetSkippedBlocksByEpoch(ctx context.Context, epoch types.Epoch) ([]types.Hash, error) {
	cfx, err := api.provider.GetClientByIP(ctx)
	if err != nil {
		return emptyHashes, err
	}

	api.inputEpochMetric.Update(&epoch, "cfx_getSkippedBlocksByEpoch", cfx)
	return cfx.GetSkippedBlocksByEpoch(&epoch)
}

func (api *cfxAPI) GetTransactionReceipt(ctx context.Context, txHash types.Hash) (*types.TransactionReceipt, error) {
	logger := logrus.WithFields(logrus.Fields{"txHash": txHash})

	if !util.IsValidHashStr(txHash.String()) {
		goto delegation
	}

	if !util.IsInterfaceValNil(api.handler) {
		isStoreHit := false
		defer func(isHit *bool) {
			hitStatsCollector.CollectHitStats("infura/rpc/call/cfx_getTransactionReceipt/store/hitratio", *isHit)
		}(&isStoreHit)

		txRcpt, err := api.handler.GetTransactionReceipt(ctx, txHash)
		if err == nil {
			logger.Debug("Loading epoch data for cfx_getTransactionReceipt hit in the store")

			isStoreHit = true
			util.StripLogExtraFieldsForRPC(txRcpt.Logs)

			return txRcpt, err
		}

		logger.WithError(err).Debug("Loading epoch data for cfx_getTransactionReceipt hit missed from the store")
	}

delegation:
	cfx, err := api.provider.GetClientByIP(ctx)
	if err != nil {
		logger.WithError(err).Debug("Failed to delegate cfx_getTransactionReceipt rpc request to fullnode")
		return nil, err
	}

	logger.WithField("fullnode", cfx.GetNodeURL()).Debug("Delegating cfx_getTransactionReceipt rpc request to fullnode")
	return cfx.GetTransactionReceipt(txHash)
}

func (api *cfxAPI) GetAccount(ctx context.Context, address types.Address, epoch *types.Epoch) (types.AccountInfo, error) {
	cfx, err := api.provider.GetClientByIP(ctx)
	if err != nil {
		return types.AccountInfo{}, err
	}

	api.inputEpochMetric.Update(epoch, "cfx_getAccount", cfx)
	return cfx.GetAccountInfo(address, toSlice(epoch)...)
}

func (api *cfxAPI) GetInterestRate(ctx context.Context, epoch *types.Epoch) (*hexutil.Big, error) {
	cfx, err := api.provider.GetClientByIP(ctx)
	if err != nil {
		return nil, err
	}

	api.inputEpochMetric.Update(epoch, "cfx_getInterestRate", cfx)
	return cfx.GetInterestRate(epoch)
}

func (api *cfxAPI) GetAccumulateInterestRate(ctx context.Context, epoch *types.Epoch) (*hexutil.Big, error) {
	cfx, err := api.provider.GetClientByIP(ctx)
	if err != nil {
		return nil, err
	}

	api.inputEpochMetric.Update(epoch, "cfx_getAccumulateInterestRate", cfx)
	return cfx.GetAccumulateInterestRate(toSlice(epoch)...)
}

func (api *cfxAPI) GetConfirmationRiskByHash(ctx context.Context, blockHash types.Hash) (*hexutil.Big, error) {
	cfx, err := api.provider.GetClientByIP(ctx)
	if err != nil {
		return nil, err
	}

	return cfx.GetRawBlockConfirmationRisk(blockHash)
}

func (api *cfxAPI) GetStatus(ctx context.Context) (types.Status, error) {
	cfx, err := api.provider.GetClientByIP(ctx)
	if err != nil {
		return types.Status{}, err
	}

	return cfx.GetStatus()
}

func (api *cfxAPI) GetBlockRewardInfo(ctx context.Context, epoch types.Epoch) ([]types.RewardInfo, error) {
	cfx, err := api.provider.GetClientByIP(ctx)
	if err != nil {
		return emptyRewards, err
	}

	api.inputEpochMetric.Update(&epoch, "cfx_getBlockRewardInfo", cfx)
	return cfx.GetBlockRewardInfo(epoch)
}

func (api *cfxAPI) ClientVersion(ctx context.Context) (string, error) {
	cfx, err := api.provider.GetClientByIP(ctx)
	if err != nil {
		return "", err
	}

	return cfx.GetClientVersion()
}

func (api *cfxAPI) GetSupplyInfo(ctx context.Context, epoch *types.Epoch) (types.TokenSupplyInfo, error) {
	cfx, err := api.provider.GetClientByIP(ctx)
	if err != nil {
		return types.TokenSupplyInfo{}, err
	}

	api.inputEpochMetric.Update(epoch, "cfx_getSupplyInfo", cfx)
	return cfx.GetSupplyInfo(toSlice(epoch)...)
}

func (api *cfxAPI) GetAccountPendingInfo(ctx context.Context, address types.Address) (*types.AccountPendingInfo, error) {
	cfx, err := api.provider.GetClientByIP(ctx)
	if err != nil {
		return nil, err
	}

	return cfx.GetAccountPendingInfo(address)
}

func (api *cfxAPI) GetAccountPendingTransactions(
	ctx context.Context, address types.Address, startNonce *hexutil.Big, limit *hexutil.Uint64,
) (*types.AccountPendingTransactions, error) {
	cfx, err := api.provider.GetClientByIP(ctx)
	if err != nil {
		return nil, err
	}

	acctPendingTxs, err := cfx.GetAccountPendingTransactions(address, startNonce, limit)
	return &acctPendingTxs, err
}

func (api *cfxAPI) GetPoSEconomics(ctx context.Context, epoch ...*types.Epoch) (*types.PoSEconomics, error) {
	cfx, err := api.provider.GetClientByIP(ctx)
	if err != nil {
		return nil, err
	}

	// TODO: remove this type assertion once Golang SDK interface method added.
	cfxc, ok := cfx.(*sdk.Client)
	if !ok {
		return nil, store.ErrUnsupported
	}

	posEconomics, err := cfxc.GetPoSEconomics(epoch...)
	return &posEconomics, err
}

// PubSub notification

// NewHeads send a notification each time a new header (block) is appended to the chain.
func (api *cfxAPI) NewHeads(ctx context.Context) (*rpc.Subscription, error) {
	psCtx, supported, err := api.pubsubCtxFromContext(ctx)

	if !supported {
		logrus.WithError(err).Error("NewHeads pubsub notification unsupported")
		return &rpc.Subscription{}, rpc.ErrNotificationsUnsupported
	}

	if err != nil {
		logrus.WithError(err).Error("NewHeads pubsub context error")
		return &rpc.Subscription{}, errSubscriptionProxyError
	}

	rpcSub := psCtx.notifier.CreateSubscription()

	headersCh := make(chan *types.BlockHeader, pubsubChannelBufferSize)
	dClient := getOrNewDelegateClient(psCtx.cfx)

	dSub, err := dClient.delegateSubscribeNewHeads(rpcSub.ID, headersCh)
	if err != nil {
		logrus.WithError(err).Error("Failed to delegate pubsub NewHeads")
		return &rpc.Subscription{}, errSubscriptionProxyError
	}

	logger := logrus.WithField("rpcSubID", rpcSub.ID)

	go func() {
		defer dSub.unsubscribe()

		for {
			select {
			case blockHeader := <-headersCh:
				logger.WithField("blockHeader", blockHeader).Debug("Received new block header from pubsub delegate")
				psCtx.notifier.Notify(rpcSub.ID, blockHeader)

			case err = <-dSub.err: // delegate subscription error
				logger.WithError(err).Debug("Received error from newHeads pubsub delegate")
				psCtx.rpcClient.Close()
				return

			case err = <-rpcSub.Err():
				logger.WithError(err).Debug("NewHeads pubsub subscription error")
				return

			case <-psCtx.notifier.Closed():
				logger.Debug("NewHeads pubsub connection closed")
				return
			}
		}
	}()

	return rpcSub, nil
}

// Epochs send a notification each time a new epoch is appended to the chain.
func (api *cfxAPI) Epochs(ctx context.Context, subEpoch *types.Epoch) (*rpc.Subscription, error) {
	if subEpoch == nil {
		subEpoch = types.EpochLatestMined
	}

	if !subEpoch.Equals(types.EpochLatestMined) && !subEpoch.Equals(types.EpochLatestState) {
		return &rpc.Subscription{}, rpc.ErrNotificationsUnsupported
	}

	psCtx, supported, err := api.pubsubCtxFromContext(ctx)
	if !supported {
		logrus.WithError(err).Errorf("Epochs pubsub notification unsupported (%v)", subEpoch)
		return &rpc.Subscription{}, rpc.ErrNotificationsUnsupported
	}

	if err != nil {
		logrus.WithError(err).Errorf("Epochs pubsub context error (%v)", subEpoch)
		return &rpc.Subscription{}, errSubscriptionProxyError
	}

	rpcSub := psCtx.notifier.CreateSubscription()

	epochsCh := make(chan *types.WebsocketEpochResponse, pubsubChannelBufferSize)
	dClient := getOrNewDelegateClient(psCtx.cfx)

	dSub, err := dClient.delegateSubscribeEpochs(rpcSub.ID, epochsCh, *subEpoch)
	if err != nil {
		logrus.WithError(err).Errorf("Failed to delegate pubsub epochs subscription (%v)", subEpoch)
		return &rpc.Subscription{}, errSubscriptionProxyError
	}

	logger := logrus.WithField("rpcSubID", rpcSub.ID)

	go func() {
		defer dSub.unsubscribe()

		for {
			select {
			case epoch := <-epochsCh:
				logger.WithField("epoch", epoch).Debugf("Received new epoch from pubsub delegate (%v)", subEpoch)
				psCtx.notifier.Notify(rpcSub.ID, epoch)

			case err = <-dSub.err: // delegate subscription error
				logger.WithError(err).Debugf("Received error from epochs pubsub delegate (%v)", subEpoch)
				psCtx.rpcClient.Close()
				return

			case err = <-rpcSub.Err():
				logger.WithError(err).Debugf("Epochs pubsub subscription error (%v)", subEpoch)
				return

			case <-psCtx.notifier.Closed():
				logger.Debugf("Epochs pubsub connection closed (%v)", subEpoch)
				return
			}
		}
	}()

	return rpcSub, nil
}

// Logs creates a subscription that fires for all new log that match the given filter criteria.
func (api *cfxAPI) Logs(ctx context.Context, filter types.LogFilter) (*rpc.Subscription, error) {
	psCtx, supported, err := api.pubsubCtxFromContext(ctx)
	if !supported {
		logrus.WithError(err).Error("Logs pubsub notification unsupported")
		return &rpc.Subscription{}, rpc.ErrNotificationsUnsupported
	}

	if err != nil {
		logrus.WithError(err).Error("Logs pubsub context error")
		return &rpc.Subscription{}, errSubscriptionProxyError
	}

	rpcSub := psCtx.notifier.CreateSubscription()

	logsCh := make(chan *types.SubscriptionLog, pubsubChannelBufferSize)
	dClient := getOrNewDelegateClient(psCtx.cfx)

	dSub, err := dClient.delegateSubscribeLogs(rpcSub.ID, logsCh, filter)
	if err != nil {
		logrus.WithField("filter", filter).WithError(err).Error("Failed to delegate pubsub logs subscription")
		return &rpc.Subscription{}, errSubscriptionProxyError
	}

	logger := logrus.WithField("rpcSubID", rpcSub.ID)

	go func() {
		defer dSub.unsubscribe()

		for {
			select {
			case log := <-logsCh:
				logger.WithField("log", log).Debug("Received new log from pubsub delegate")
				psCtx.notifier.Notify(rpcSub.ID, log)

			case err = <-dSub.err: // delegate subscription error
				logger.WithError(err).Debug("Received error from logs pubsub delegate")
				psCtx.rpcClient.Close()
				return

			case err = <-rpcSub.Err():
				logger.WithError(err).Debugf("Logs pubsub subscription error")
				return

			case <-psCtx.notifier.Closed():
				logger.Debugf("Logs pubsub connection closed")
				return
			}
		}
	}()

	return rpcSub, nil
}

type pubsubContext struct {
	notifier  *rpc.Notifier
	rpcClient *rpc.Client
	cfx       sdk.ClientOperator
}

// pubsubCtxFromContext returns the pubsub context with member variables stored in ctx, if any.
func (api *cfxAPI) pubsubCtxFromContext(ctx context.Context) (psCtx *pubsubContext, supported bool, err error) {
	notifier, supported := rpc.NotifierFromContext(ctx)
	if !supported {
		err = errors.New("failed to get notifier from context")
		return
	}

	rpcClient, supported := rpcClientFromContext(ctx)
	if !supported {
		err = errors.New("failed to get rpc client from context")
		return
	}

	cfx, err := api.provider.GetWSClientByIP(ctx)
	if err != nil {
		err = errors.WithMessage(err, "failed to get cfx wsclient by ip")
		return
	}

	psCtx = &pubsubContext{notifier, rpcClient, cfx}
	return
}