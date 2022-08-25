package mainchannel

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-config/configtx"
	"github.com/hyperledger/fabric-config/configtx/membership"
	"github.com/hyperledger/fabric-config/configtx/orderer"
	cb "github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite/bccsp/sw"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/resource"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric/protoutil"
	hlfv1alpha1 "github.com/kfsoftware/hlf-operator/api/hlf.kungfusoftware.es/v1alpha1"
	"github.com/kfsoftware/hlf-operator/controllers/testutils"
	"github.com/kfsoftware/hlf-operator/controllers/utils"
	"github.com/kfsoftware/hlf-operator/internal/github.com/hyperledger/fabric-ca/sdkinternal/pkg/util"
	"github.com/kfsoftware/hlf-operator/kubectl-hlf/cmd/helpers"
	"github.com/kfsoftware/hlf-operator/kubectl-hlf/cmd/helpers/osnadmin"
	operatorv1 "github.com/kfsoftware/hlf-operator/pkg/client/clientset/versioned"
	"github.com/kfsoftware/hlf-operator/pkg/nc"
	"github.com/operator-framework/operator-lib/status"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"strings"
	"time"
)

// FabricMainChannelReconciler reconciles a FabricMainChannel object
type FabricMainChannelReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	Config *rest.Config
}

const mainChannelFinalizer = "finalizer.mainChannel.hlf.kungfusoftware.es"

func (r *FabricMainChannelReconciler) finalizeMainChannel(reqLogger logr.Logger, m *hlfv1alpha1.FabricMainChannel) error {
	ns := m.Namespace
	if ns == "" {
		ns = "default"
	}
	//releaseName := m.Name
	reqLogger.Info("Successfully finalized mainChannel")

	return nil
}

func (r *FabricMainChannelReconciler) addFinalizer(reqLogger logr.Logger, m *hlfv1alpha1.FabricMainChannel) error {
	reqLogger.Info("Adding Finalizer for the MainChannel")
	controllerutil.AddFinalizer(m, mainChannelFinalizer)

	// Update CR
	err := r.Update(context.TODO(), m)
	if err != nil {
		reqLogger.Error(err, "Failed to update MainChannel with finalizer")
		return err
	}
	return nil
}

// +kubebuilder:rbac:groups=hlf.kungfusoftware.es,resources=fabricmainchannels,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=hlf.kungfusoftware.es,resources=fabricmainchannels/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=hlf.kungfusoftware.es,resources=fabricmainchannels/finalizers,verbs=get;update;patch
func (r *FabricMainChannelReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	reqLogger := r.Log.WithValues("hlf", req.NamespacedName)
	fabricMainChannel := &hlfv1alpha1.FabricMainChannel{}

	err := r.Get(ctx, req.NamespacedName, fabricMainChannel)
	if err != nil {
		log.Debugf("Error getting the object %s error=%v", req.NamespacedName, err)
		if apierrors.IsNotFound(err) {
			reqLogger.Info("MainChannel resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		reqLogger.Error(err, "Failed to get MainChannel.")
		return ctrl.Result{}, err
	}
	clientSet, err := utils.GetClientKubeWithConf(r.Config)
	if err != nil {
		r.setConditionStatus(ctx, fabricMainChannel, hlfv1alpha1.FailedStatus, false, err, false)
		return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricMainChannel)
	}
	hlfClientSet, err := operatorv1.NewForConfig(r.Config)
	if err != nil {
		r.setConditionStatus(ctx, fabricMainChannel, hlfv1alpha1.FailedStatus, false, err, false)
		return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricMainChannel)
	}
	chStore := testutils.NewChannelStore()
	var consenters []testutils.Consenter
	for _, consenter := range fabricMainChannel.Spec.Consenters {
		tlsCert, err := utils.ParseX509Certificate([]byte(consenter.TLSCert))
		if err != nil {
			r.setConditionStatus(ctx, fabricMainChannel, hlfv1alpha1.FailedStatus, false, err, false)
			return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricMainChannel)
		}
		channelConsenter := testutils.CreateConsenter(
			consenter.Host,
			consenter.Port,
			tlsCert,
		)
		consenters = append(consenters, channelConsenter)
	}
	var ordererOrgs []testutils.OrdererOrg
	for _, ordererOrg := range fabricMainChannel.Spec.OrdererOrganizations {
		certAuth, err := helpers.GetCertAuthByName(
			clientSet,
			hlfClientSet,
			ordererOrg.CAName,
			ordererOrg.CANamespace,
		)
		if err != nil {
			r.setConditionStatus(ctx, fabricMainChannel, hlfv1alpha1.FailedStatus, false, err, false)
			return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricMainChannel)
		}
		tlsCACert, err := utils.ParseX509Certificate([]byte(certAuth.Status.TLSCACert))
		if err != nil {
			r.setConditionStatus(ctx, fabricMainChannel, hlfv1alpha1.FailedStatus, false, err, false)
			return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricMainChannel)
		}
		caCert, err := utils.ParseX509Certificate([]byte(certAuth.Status.CACert))
		if err != nil {
			r.setConditionStatus(ctx, fabricMainChannel, hlfv1alpha1.FailedStatus, false, err, false)
			return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricMainChannel)
		}
		ordererOrgs = append(ordererOrgs, testutils.CreateOrdererOrg(
			ordererOrg.MSPID,
			tlsCACert,
			caCert,
			ordererOrg.OrdererEndpoints,
		))
	}
	for _, ordererOrg := range fabricMainChannel.Spec.ExternalOrdererOrganizations {
		rootCert, err := utils.ParseX509Certificate([]byte(ordererOrg.SignRootCert))
		if err != nil {
			r.setConditionStatus(ctx, fabricMainChannel, hlfv1alpha1.FailedStatus, false, err, false)
			return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricMainChannel)
		}
		tlsRootCert, err := utils.ParseX509Certificate([]byte(ordererOrg.TLSRootCert))
		if err != nil {
			r.setConditionStatus(ctx, fabricMainChannel, hlfv1alpha1.FailedStatus, false, err, false)
			return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricMainChannel)
		}
		ordererOrgs = append(ordererOrgs, testutils.CreateOrdererOrg(
			ordererOrg.MSPID,
			tlsRootCert,
			rootCert,
			ordererOrg.OrdererEndpoints,
		))
	}
	var peerOrgs []testutils.PeerOrg

	for _, peerOrg := range fabricMainChannel.Spec.PeerOrganizations {
		certAuth, err := helpers.GetCertAuthByName(
			clientSet,
			hlfClientSet,
			peerOrg.CAName,
			peerOrg.CANamespace,
		)
		if err != nil {
			r.setConditionStatus(ctx, fabricMainChannel, hlfv1alpha1.FailedStatus, false, err, false)
			return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricMainChannel)
		}
		tlsCACert, err := utils.ParseX509Certificate([]byte(certAuth.Status.TLSCACert))
		if err != nil {
			r.setConditionStatus(ctx, fabricMainChannel, hlfv1alpha1.FailedStatus, false, err, false)
			return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricMainChannel)
		}
		caCert, err := utils.ParseX509Certificate([]byte(certAuth.Status.CACert))
		if err != nil {
			r.setConditionStatus(ctx, fabricMainChannel, hlfv1alpha1.FailedStatus, false, err, false)
			return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricMainChannel)
		}
		peerOrgs = append(peerOrgs, testutils.CreatePeerOrg(
			peerOrg.MSPID,
			tlsCACert,
			caCert,
		))
	}
	for _, peerOrg := range fabricMainChannel.Spec.ExternalPeerOrganizations {
		rootCert, err := utils.ParseX509Certificate([]byte(peerOrg.SignRootCert))
		if err != nil {
			r.setConditionStatus(ctx, fabricMainChannel, hlfv1alpha1.FailedStatus, false, err, false)
			return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricMainChannel)
		}
		tlsRootCert, err := utils.ParseX509Certificate([]byte(peerOrg.TLSRootCert))
		if err != nil {
			r.setConditionStatus(ctx, fabricMainChannel, hlfv1alpha1.FailedStatus, false, err, false)
			return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricMainChannel)
		}
		peerOrgs = append(peerOrgs, testutils.CreatePeerOrg(
			peerOrg.MSPID,
			tlsRootCert,
			rootCert,
		))
	}
	channelOptions := []testutils.ChannelOption{
		testutils.WithName(fabricMainChannel.Spec.Name),
		testutils.WithOrdererOrgs(ordererOrgs...),
		testutils.WithPeerOrgs(peerOrgs...),
		testutils.WithConsenters(consenters...),
	}
	channelOptions = append(
		channelOptions,
		testutils.WithBatchSize(&orderer.BatchSize{
			MaxMessageCount:   uint32(fabricMainChannel.Spec.ChannelConfig.Orderer.BatchSize.MaxMessageCount),
			AbsoluteMaxBytes:  uint32(fabricMainChannel.Spec.ChannelConfig.Orderer.BatchSize.AbsoluteMaxBytes),
			PreferredMaxBytes: uint32(fabricMainChannel.Spec.ChannelConfig.Orderer.BatchSize.PreferredMaxBytes),
		}),
	)
	block, err := chStore.GetApplicationChannelBlock(
		ctx,
		channelOptions...,
	)
	if err != nil {
		r.setConditionStatus(ctx, fabricMainChannel, hlfv1alpha1.FailedStatus, false, err, false)
		return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricMainChannel)
	}
	blockBytes, err := proto.Marshal(block)
	if err != nil {
		r.setConditionStatus(ctx, fabricMainChannel, hlfv1alpha1.FailedStatus, false, err, false)
		return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricMainChannel)
	}
	// join orderers
	for _, ordererOrg := range fabricMainChannel.Spec.OrdererOrganizations {
		certAuth, err := helpers.GetCertAuthByName(
			clientSet,
			hlfClientSet,
			ordererOrg.CAName,
			ordererOrg.CANamespace,
		)
		if err != nil {
			r.setConditionStatus(ctx, fabricMainChannel, hlfv1alpha1.FailedStatus, false, err, false)
			return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricMainChannel)
		}
		certPool := x509.NewCertPool()
		ok := certPool.AppendCertsFromPEM([]byte(certAuth.Status.TLSCACert))
		if !ok {
			r.setConditionStatus(ctx, fabricMainChannel, hlfv1alpha1.FailedStatus, false, fmt.Errorf(" %s", ordererOrg.MSPID), false)
			return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricMainChannel)
		}
		idConfig, ok := fabricMainChannel.Spec.Identities[ordererOrg.MSPID]
		if !ok {
			r.setConditionStatus(ctx, fabricMainChannel, hlfv1alpha1.FailedStatus, false, fmt.Errorf("identity not found for MSPID %s", ordererOrg.MSPID), false)
			return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricMainChannel)
		}
		secret, err := clientSet.CoreV1().Secrets(fabricMainChannel.Namespace).Get(ctx, idConfig.SecretName, v1.GetOptions{})
		if err != nil {
			r.setConditionStatus(ctx, fabricMainChannel, hlfv1alpha1.FailedStatus, false, err, false)
			return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricMainChannel)
		}
		id := &identity{}
		secretData, ok := secret.Data[idConfig.SecretKey]
		if !ok {
			r.setConditionStatus(ctx, fabricMainChannel, hlfv1alpha1.FailedStatus, false, fmt.Errorf("secret key %s not found", idConfig.SecretKey), false)
			return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricMainChannel)
		}
		err = yaml.Unmarshal(secretData, id)
		if err != nil {
			r.setConditionStatus(ctx, fabricMainChannel, hlfv1alpha1.FailedStatus, false, err, false)
			return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricMainChannel)
		}
		tlsClientCert, err := tls.X509KeyPair(
			[]byte(id.Cert.Pem),
			[]byte(id.Key.Pem),
		)
		for _, cc := range ordererOrg.ExternalOrderersToJoin {
			osnUrl := fmt.Sprintf("https://%s:%d", cc.Host, cc.AdminPort)
			chResponse, err := osnadmin.Join(osnUrl, blockBytes, certPool, tlsClientCert)
			if err != nil {
				r.setConditionStatus(ctx, fabricMainChannel, hlfv1alpha1.FailedStatus, false, err, false)
				return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricMainChannel)
			}
			defer chResponse.Body.Close()
			log.Infof("Orderer %s joined Status code=%d", osnUrl, chResponse.StatusCode)
			if chResponse.StatusCode != 201 {
				r.setConditionStatus(ctx, fabricMainChannel, hlfv1alpha1.FailedStatus, false, err, false)
				return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricMainChannel)
			}
			chInfo := &osnadmin.ChannelInfo{}
			err = json.NewDecoder(chResponse.Body).Decode(chInfo)
			if err != nil {
				r.setConditionStatus(ctx, fabricMainChannel, hlfv1alpha1.FailedStatus, false, err, false)
				return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricMainChannel)
			}
		}

		for _, cc := range ordererOrg.OrderersToJoin {
			adminPort := 9444
			osnUrl := fmt.Sprintf("https://%s.%s:%d", cc.Name, cc.Namespace, adminPort)
			chResponse, err := osnadmin.Join(osnUrl, blockBytes, certPool, tlsClientCert)
			if err != nil {
				r.setConditionStatus(ctx, fabricMainChannel, hlfv1alpha1.FailedStatus, false, err, false)
				return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricMainChannel)
			}
			defer chResponse.Body.Close()
			log.Infof("Orderer %s.%s joined Status code=%d", cc.Name, cc.Namespace, chResponse.StatusCode)
			if chResponse.StatusCode != 201 {
				r.setConditionStatus(ctx, fabricMainChannel, hlfv1alpha1.FailedStatus, false, err, false)
				return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricMainChannel)
			}
			chInfo := &osnadmin.ChannelInfo{}
			err = json.NewDecoder(chResponse.Body).Decode(chInfo)
			if err != nil {
				r.setConditionStatus(ctx, fabricMainChannel, hlfv1alpha1.FailedStatus, false, err, false)
				return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricMainChannel)
			}
		}
	}

	ncResponse, err := nc.GenerateNetworkConfig(clientSet, hlfClientSet, "")
	if err != nil {
		r.setConditionStatus(ctx, fabricMainChannel, hlfv1alpha1.FailedStatus, false, errors.Wrapf(err, "failed to generate network config"), false)
		return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricMainChannel)
	}
	configBackend := config.FromRaw([]byte(ncResponse.NetworkConfig), "yaml")
	sdk, err := fabsdk.New(configBackend)
	if err != nil {
		r.setConditionStatus(ctx, fabricMainChannel, hlfv1alpha1.FailedStatus, false, err, false)
		return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricMainChannel)
	}
	firstAdminOrgMSPID := fabricMainChannel.Spec.AdminPeerOrganizations[0].MSPID
	idConfig, ok := fabricMainChannel.Spec.Identities[firstAdminOrgMSPID]
	if !ok {
		r.setConditionStatus(ctx, fabricMainChannel, hlfv1alpha1.FailedStatus, false, fmt.Errorf("identity not found for MSPID %s", firstAdminOrgMSPID), false)
		return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricMainChannel)
	}
	secret, err := clientSet.CoreV1().Secrets(fabricMainChannel.Namespace).Get(ctx, idConfig.SecretName, v1.GetOptions{})
	if err != nil {
		r.setConditionStatus(ctx, fabricMainChannel, hlfv1alpha1.FailedStatus, false, err, false)
		return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricMainChannel)
	}
	secretData, ok := secret.Data[idConfig.SecretKey]
	if !ok {
		r.setConditionStatus(ctx, fabricMainChannel, hlfv1alpha1.FailedStatus, false, fmt.Errorf("secret key %s not found", idConfig.SecretKey), false)
		return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricMainChannel)
	}
	id := &identity{}
	err = yaml.Unmarshal(secretData, id)
	if err != nil {
		r.setConditionStatus(ctx, fabricMainChannel, hlfv1alpha1.FailedStatus, false, err, false)
		return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricMainChannel)
	}
	cert, err := utils.ParseX509Certificate([]byte(id.Cert.Pem))
	if err != nil {
		r.setConditionStatus(ctx, fabricMainChannel, hlfv1alpha1.FailedStatus, false, err, false)
		return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricMainChannel)
	}
	sdkConfig, err := sdk.Config()
	if err != nil {
		r.setConditionStatus(ctx, fabricMainChannel, hlfv1alpha1.FailedStatus, false, err, false)
		return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricMainChannel)
	}
	cryptoConfig := cryptosuite.ConfigFromBackend(sdkConfig)
	cryptoSuite, err := sw.GetSuiteByConfig(cryptoConfig)
	if err != nil {
		r.setConditionStatus(ctx, fabricMainChannel, hlfv1alpha1.FailedStatus, false, err, false)
		return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricMainChannel)
	}
	coreKey, err := util.ImportBCCSPKeyFromPEM(id.Key.Pem, cryptoSuite, true)
	if err != nil {
		r.setConditionStatus(ctx, fabricMainChannel, hlfv1alpha1.FailedStatus, false, err, false)
		return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricMainChannel)
	}
	fabricIdentity := fabricUser{
		x509Cert:   cert,
		privateKey: coreKey,
	}
	sdkContext := sdk.Context(
		fabsdk.WithIdentity(fabricIdentity),
		fabsdk.WithOrg(firstAdminOrgMSPID),
	)
	resClient, err := resmgmt.New(sdkContext)
	if err != nil {
		r.setConditionStatus(ctx, fabricMainChannel, hlfv1alpha1.FailedStatus, false, err, false)
		return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricMainChannel)
	}
	block1, err := resClient.QueryConfigBlockFromOrderer(fabricMainChannel.Spec.Name)
	if err != nil {
		log.Infof("channel %s does not exist, it will be created", fabricMainChannel.Spec.Name)
	}
	cfgBlock, err := resource.ExtractConfigFromBlock(block1)
	if err != nil {
		r.setConditionStatus(ctx, fabricMainChannel, hlfv1alpha1.FailedStatus, false, errors.Wrapf(err, "failed to extract config from channel block"), false)
		return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricMainChannel)
	}
	updatedConfigTX := configtx.New(cfgBlock)
	configTX, err := r.mapToConfigTX(fabricMainChannel)
	if err != nil {
		r.setConditionStatus(ctx, fabricMainChannel, hlfv1alpha1.FailedStatus, false, errors.Wrapf(err, "error mapping channel to configtx channel"), false)
		return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricMainChannel)
	}
	err = updateApplicationChannelConfigTx(updatedConfigTX, configTX)
	if err != nil {
		r.setConditionStatus(ctx, fabricMainChannel, hlfv1alpha1.FailedStatus, false, errors.Wrapf(err, "failed to extract config from channel block"), false)
		return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricMainChannel)
	}
	configUpdate, err := resmgmt.CalculateConfigUpdate(fabricMainChannel.Spec.Name, cfgBlock, updatedConfigTX.UpdatedConfig())
	if err != nil {
		if !strings.Contains(err.Error(), "no differences detected between original and updated config") {
			r.setConditionStatus(ctx, fabricMainChannel, hlfv1alpha1.FailedStatus, false, errors.Wrapf(err, "error calculating config update"), false)
			return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricMainChannel)
		}
		log.Infof("No differences detected between original and updated config")
	} else {
		channelConfigBytes, err := CreateConfigUpdateEnvelope(fabricMainChannel.Spec.Name, configUpdate)
		if err != nil {
			r.setConditionStatus(ctx, fabricMainChannel, hlfv1alpha1.FailedStatus, false, errors.Wrapf(err, "error creating config update envelope"), false)
			return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricMainChannel)
		}
		var configSignatures []*cb.ConfigSignature
		for _, adminPeer := range fabricMainChannel.Spec.AdminPeerOrganizations {
			configUpdateReader := bytes.NewReader(channelConfigBytes)
			idConfig, ok := fabricMainChannel.Spec.Identities[adminPeer.MSPID]
			if !ok {
				r.setConditionStatus(ctx, fabricMainChannel, hlfv1alpha1.FailedStatus, false, fmt.Errorf("identity not found for MSPID %s", adminPeer.MSPID), false)
				return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricMainChannel)
			}
			secret, err := clientSet.CoreV1().Secrets(fabricMainChannel.Namespace).Get(ctx, idConfig.SecretName, v1.GetOptions{})
			if err != nil {
				r.setConditionStatus(ctx, fabricMainChannel, hlfv1alpha1.FailedStatus, false, err, false)
				return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricMainChannel)
			}
			secretData, ok := secret.Data[idConfig.SecretKey]
			if !ok {
				r.setConditionStatus(ctx, fabricMainChannel, hlfv1alpha1.FailedStatus, false, fmt.Errorf("secret key %s not found", idConfig.SecretKey), false)
				return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricMainChannel)
			}
			id := &identity{}
			err = yaml.Unmarshal(secretData, id)
			if err != nil {
				r.setConditionStatus(ctx, fabricMainChannel, hlfv1alpha1.FailedStatus, false, err, false)
				return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricMainChannel)
			}
			cert, err := utils.ParseX509Certificate([]byte(id.Cert.Pem))
			if err != nil {
				r.setConditionStatus(ctx, fabricMainChannel, hlfv1alpha1.FailedStatus, false, err, false)
				return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricMainChannel)
			}
			sdkConfig, err := sdk.Config()
			if err != nil {
				r.setConditionStatus(ctx, fabricMainChannel, hlfv1alpha1.FailedStatus, false, err, false)
				return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricMainChannel)
			}
			cryptoConfig := cryptosuite.ConfigFromBackend(sdkConfig)
			cryptoSuite, err := sw.GetSuiteByConfig(cryptoConfig)
			if err != nil {
				r.setConditionStatus(ctx, fabricMainChannel, hlfv1alpha1.FailedStatus, false, err, false)
				return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricMainChannel)
			}
			coreKey, err := util.ImportBCCSPKeyFromPEM(id.Key.Pem, cryptoSuite, true)
			if err != nil {
				r.setConditionStatus(ctx, fabricMainChannel, hlfv1alpha1.FailedStatus, false, err, false)
				return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricMainChannel)
			}
			fabricIdentity := fabricUser{
				x509Cert:   cert,
				privateKey: coreKey,
			}
			sdkContext := sdk.Context(
				fabsdk.WithIdentity(fabricIdentity),
				fabsdk.WithOrg(adminPeer.MSPID),
			)
			resClient, err := resmgmt.New(sdkContext)
			if err != nil {
				r.setConditionStatus(ctx, fabricMainChannel, hlfv1alpha1.FailedStatus, false, err, false)
				return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricMainChannel)
			}
			signature, err := resClient.CreateConfigSignatureFromReader(fabricIdentity, configUpdateReader)
			if err != nil {
				r.setConditionStatus(ctx, fabricMainChannel, hlfv1alpha1.FailedStatus, false, err, false)
				return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricMainChannel)
			}
			configSignatures = append(configSignatures, signature)
		}
		configUpdateReader := bytes.NewReader(channelConfigBytes)
		saveChannelResponse, err := resClient.SaveChannel(
			resmgmt.SaveChannelRequest{
				ChannelID:         fabricMainChannel.Spec.Name,
				ChannelConfig:     configUpdateReader,
				SigningIdentities: []msp.SigningIdentity{},
			},
			resmgmt.WithConfigSignatures(configSignatures...),
		)
		if err != nil {
			r.setConditionStatus(ctx, fabricMainChannel, hlfv1alpha1.FailedStatus, false, errors.Wrapf(err, "error saving application configuration"), false)
			return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricMainChannel)
		}
		log.Infof("Application configuration updated with transaction ID: %s", saveChannelResponse.TransactionID)
	}

	isMainChannelMarkedToBeDeleted := fabricMainChannel.GetDeletionTimestamp() != nil
	if isMainChannelMarkedToBeDeleted {
		if utils.Contains(fabricMainChannel.GetFinalizers(), mainChannelFinalizer) {
			if err := r.finalizeMainChannel(reqLogger, fabricMainChannel); err != nil {
				return ctrl.Result{}, err
			}
			controllerutil.RemoveFinalizer(fabricMainChannel, mainChannelFinalizer)
			err := r.Update(ctx, fabricMainChannel)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}
	if !utils.Contains(fabricMainChannel.GetFinalizers(), mainChannelFinalizer) {
		if err := r.addFinalizer(reqLogger, fabricMainChannel); err != nil {
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

type fabricUser struct {
	msp.SigningIdentity
	privateKey core.Key
	x509Cert   *x509.Certificate
	id         string
	mspID      string
}

func (f fabricUser) Identifier() *msp.IdentityIdentifier {
	return &msp.IdentityIdentifier{MSPID: f.mspID, ID: f.id}
}

func (f fabricUser) Verify(msg []byte, sig []byte) error {
	return errors.New("not implemented")
}

func (f fabricUser) Serialize() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (f fabricUser) EnrollmentCertificate() []byte {
	return utils.EncodeX509Certificate(f.x509Cert)
}

func (f fabricUser) Sign(msg []byte) ([]byte, error) {
	return nil, errors.New("Sign() function not implemented")
}

func (f fabricUser) PublicVersion() msp.Identity {
	return f
}

func (f fabricUser) PrivateKey() core.Key {
	return f.privateKey
}

var (
	ErrClientK8s = errors.New("k8sAPIClientError")
)

func (r *FabricMainChannelReconciler) updateCRStatusOrFailReconcile(ctx context.Context, log logr.Logger, p *hlfv1alpha1.FabricMainChannel) (
	reconcile.Result, error) {
	if err := r.Status().Update(ctx, p); err != nil {
		log.Error(err, fmt.Sprintf("%v failed to update the application status", ErrClientK8s))
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

func (r *FabricMainChannelReconciler) setConditionStatus(ctx context.Context, p *hlfv1alpha1.FabricMainChannel, conditionType hlfv1alpha1.DeploymentStatus, statusFlag bool, err error, statusUnknown bool) (update bool) {
	statusStr := func() corev1.ConditionStatus {
		if statusUnknown {
			return corev1.ConditionUnknown
		}
		if statusFlag {
			return corev1.ConditionTrue
		} else {
			return corev1.ConditionFalse
		}
	}
	if p.Status.Status != conditionType {
		depCopy := client.MergeFrom(p.DeepCopy())
		p.Status.Status = conditionType
		err = r.Status().Patch(ctx, p, depCopy)
		if err != nil {
			log.Warnf("Failed to update status to %s: %v", conditionType, err)
		}
	}
	if err != nil {
		p.Status.Message = err.Error()
	}
	condition := func() status.Condition {
		if err != nil {
			return status.Condition{
				Type:    status.ConditionType(conditionType),
				Status:  statusStr(),
				Reason:  status.ConditionReason(err.Error()),
				Message: err.Error(),
			}
		}
		return status.Condition{
			Type:   status.ConditionType(conditionType),
			Status: statusStr(),
		}
	}
	return p.Status.Conditions.SetCondition(condition())
}

func (r *FabricMainChannelReconciler) SetupWithManager(mgr ctrl.Manager) error {
	managedBy := ctrl.NewControllerManagedBy(mgr)
	return managedBy.
		For(&hlfv1alpha1.FabricMainChannel{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}

func (r *FabricMainChannelReconciler) mapToConfigTX(channel *hlfv1alpha1.FabricMainChannel) (configtx.Channel, error) {
	var consenters []orderer.Consenter
	for _, consenter := range channel.Spec.Consenters {
		tlsCert, err := utils.ParseX509Certificate([]byte(consenter.TLSCert))
		if err != nil {
			return configtx.Channel{}, err
		}
		channelConsenter := orderer.Consenter{
			Address: orderer.EtcdAddress{
				Host: consenter.Host,
				Port: consenter.Port,
			},
			ClientTLSCert: tlsCert,
			ServerTLSCert: tlsCert,
		}
		consenters = append(consenters, channelConsenter)
	}
	clientSet, err := utils.GetClientKubeWithConf(r.Config)
	if err != nil {
		return configtx.Channel{}, err
	}
	hlfClientSet, err := operatorv1.NewForConfig(r.Config)
	if err != nil {
		return configtx.Channel{}, err
	}
	var ordererOrgs []configtx.Organization
	for _, ordererOrg := range channel.Spec.OrdererOrganizations {
		certAuth, err := helpers.GetCertAuthByName(
			clientSet,
			hlfClientSet,
			ordererOrg.CAName,
			ordererOrg.CANamespace,
		)
		if err != nil {
			return configtx.Channel{}, err
		}
		tlsCACert, err := utils.ParseX509Certificate([]byte(certAuth.Status.TLSCACert))
		if err != nil {
			return configtx.Channel{}, err
		}
		caCert, err := utils.ParseX509Certificate([]byte(certAuth.Status.CACert))
		if err != nil {
			return configtx.Channel{}, err
		}
		ordererOrgs = append(ordererOrgs, configtx.Organization{
			Name:     ordererOrg.MSPID,
			Policies: map[string]configtx.Policy{},
			MSP: configtx.MSP{
				Name:         ordererOrg.MSPID,
				RootCerts:    []*x509.Certificate{caCert},
				TLSRootCerts: []*x509.Certificate{tlsCACert},
				NodeOUs: membership.NodeOUs{
					Enable: true,
					ClientOUIdentifier: membership.OUIdentifier{
						Certificate:                  caCert,
						OrganizationalUnitIdentifier: "client",
					},
					PeerOUIdentifier: membership.OUIdentifier{
						Certificate:                  caCert,
						OrganizationalUnitIdentifier: "peer",
					},
					AdminOUIdentifier: membership.OUIdentifier{
						Certificate:                  caCert,
						OrganizationalUnitIdentifier: "admin",
					},
					OrdererOUIdentifier: membership.OUIdentifier{
						Certificate:                  caCert,
						OrganizationalUnitIdentifier: "orderer",
					},
				},
				Admins:                        []*x509.Certificate{},
				IntermediateCerts:             []*x509.Certificate{},
				RevocationList:                []*pkix.CertificateList{},
				OrganizationalUnitIdentifiers: []membership.OUIdentifier{},
				CryptoConfig:                  membership.CryptoConfig{},
				TLSIntermediateCerts:          []*x509.Certificate{},
			},
			AnchorPeers:      []configtx.Address{},
			OrdererEndpoints: ordererOrg.OrdererEndpoints,
			ModPolicy:        "",
		})
	}
	ordConfigtx := configtx.Orderer{
		OrdererType:   "etcdraft",
		Organizations: ordererOrgs,
		EtcdRaft: orderer.EtcdRaft{
			Consenters: consenters,
			Options: orderer.EtcdRaftOptions{
				TickInterval:         "500ms",
				ElectionTick:         10,
				HeartbeatTick:        1,
				MaxInflightBlocks:    5,
				SnapshotIntervalSize: 16 * 1024 * 1024, // 16 MB
			},
		},
		Policies: map[string]configtx.Policy{
			"Readers": {
				Type: "ImplicitMeta",
				Rule: "ANY Readers",
			},
			"Writers": {
				Type: "ImplicitMeta",
				Rule: "ANY Writers",
			},
			"Admins": {
				Type: "ImplicitMeta",
				Rule: "MAJORITY Admins",
			},
			"BlockValidation": {
				Type: "ImplicitMeta",
				Rule: "ANY Writers",
			},
		},
		Capabilities: []string{"V2_0"},
		BatchSize: orderer.BatchSize{
			MaxMessageCount:   100,
			AbsoluteMaxBytes:  1024 * 1024,
			PreferredMaxBytes: 512 * 1024,
		},
		BatchTimeout: 2 * time.Second,
		State:        "STATE_NORMAL",
	}
	var peerOrgs []configtx.Organization
	for _, peerOrg := range channel.Spec.PeerOrganizations {
		certAuth, err := helpers.GetCertAuthByName(
			clientSet,
			hlfClientSet,
			peerOrg.CAName,
			peerOrg.CANamespace,
		)
		if err != nil {
			return configtx.Channel{}, err
		}
		tlsCACert, err := utils.ParseX509Certificate([]byte(certAuth.Status.TLSCACert))
		if err != nil {
			return configtx.Channel{}, err
		}
		caCert, err := utils.ParseX509Certificate([]byte(certAuth.Status.CACert))
		if err != nil {
			return configtx.Channel{}, err
		}
		ordererOrgs = append(ordererOrgs, configtx.Organization{
			Name:     peerOrg.MSPID,
			Policies: map[string]configtx.Policy{},
			MSP: configtx.MSP{
				Name:         peerOrg.MSPID,
				RootCerts:    []*x509.Certificate{caCert},
				TLSRootCerts: []*x509.Certificate{tlsCACert},
				NodeOUs: membership.NodeOUs{
					Enable: true,
					ClientOUIdentifier: membership.OUIdentifier{
						Certificate:                  caCert,
						OrganizationalUnitIdentifier: "client",
					},
					PeerOUIdentifier: membership.OUIdentifier{
						Certificate:                  caCert,
						OrganizationalUnitIdentifier: "peer",
					},
					AdminOUIdentifier: membership.OUIdentifier{
						Certificate:                  caCert,
						OrganizationalUnitIdentifier: "admin",
					},
					OrdererOUIdentifier: membership.OUIdentifier{
						Certificate:                  caCert,
						OrganizationalUnitIdentifier: "orderer",
					},
				},
				Admins:                        []*x509.Certificate{},
				IntermediateCerts:             []*x509.Certificate{},
				RevocationList:                []*pkix.CertificateList{},
				OrganizationalUnitIdentifiers: []membership.OUIdentifier{},
				CryptoConfig:                  membership.CryptoConfig{},
				TLSIntermediateCerts:          []*x509.Certificate{},
			},
			AnchorPeers:      []configtx.Address{},
			OrdererEndpoints: []string{},
			ModPolicy:        "",
		})
	}
	application := configtx.Application{
		Organizations: peerOrgs,
		Capabilities:  []string{"V2_0"},
		Policies: map[string]configtx.Policy{
			"Readers": {
				Type: "ImplicitMeta",
				Rule: "ANY Readers",
			},
			"Writers": {
				Type: "ImplicitMeta",
				Rule: "ANY Writers",
			},
			"Admins": {
				Type: "ImplicitMeta",
				Rule: "MAJORITY Admins",
			},
			"Endorsement": {
				Type: "ImplicitMeta",
				Rule: "MAJORITY Endorsement",
			},
			"LifecycleEndorsement": {
				Type: "ImplicitMeta",
				Rule: "MAJORITY Endorsement",
			},
		},
		ACLs: defaultACLs(),
	}
	channelConfig := configtx.Channel{
		Orderer:      ordConfigtx,
		Application:  application,
		Capabilities: []string{"V2_0"},
		Policies: map[string]configtx.Policy{
			"Readers": {
				Type: "ImplicitMeta",
				Rule: "ANY Readers",
			},
			"Writers": {
				Type: "ImplicitMeta",
				Rule: "ANY Writers",
			},
			"Admins": {
				Type: "ImplicitMeta",
				Rule: "MAJORITY Admins",
			},
		},
	}
	return channelConfig, nil
}

type identity struct {
	Cert Pem `json:"cert"`
	Key  Pem `json:"key"`
}
type Pem struct {
	Pem string
}

func CreateConfigUpdateEnvelope(channelID string, configUpdate *cb.ConfigUpdate) ([]byte, error) {
	configUpdate.ChannelId = channelID
	configUpdateData, err := proto.Marshal(configUpdate)
	if err != nil {
		return nil, err
	}
	configUpdateEnvelope := &cb.ConfigUpdateEnvelope{}
	configUpdateEnvelope.ConfigUpdate = configUpdateData
	envelope, err := protoutil.CreateSignedEnvelope(cb.HeaderType_CONFIG_UPDATE, channelID, nil, configUpdateEnvelope, 0, 0)
	if err != nil {
		return nil, err
	}
	envelopeData, err := proto.Marshal(envelope)
	if err != nil {
		return nil, err
	}
	return envelopeData, nil
}

func updateApplicationChannelConfigTx(currentConfigTX configtx.ConfigTx, newConfigTx configtx.Channel) error {
	err := currentConfigTX.Application().SetPolicies(
		newConfigTx.Application.Policies,
	)
	if err != nil {
		return errors.Wrapf(err, "failed to set application")
	}
	app, err := currentConfigTX.Application().Configuration()
	if err != nil {
		return errors.Wrapf(err, "failed to get application configuration")
	}

	for _, channelPeerOrg := range app.Organizations {
		deleted := true
		for _, organization := range newConfigTx.Application.Organizations {
			if organization.Name == channelPeerOrg.Name {
				deleted = false
				break
			}
		}
		if deleted {
			currentConfigTX.Application().RemoveOrganization(channelPeerOrg.Name)
		}
	}
	err = currentConfigTX.Application().SetACLs(
		newConfigTx.Application.ACLs,
	)
	if err != nil {
		return errors.Wrapf(err, "failed to set ACLs")
	}
	return nil
}

func defaultACLs() map[string]string {
	return map[string]string{
		"_lifecycle/CheckCommitReadiness": "/Channel/Application/Writers",

		//  ACL policy for _lifecycle's "CommitChaincodeDefinition" function
		"_lifecycle/CommitChaincodeDefinition": "/Channel/Application/Writers",

		//  ACL policy for _lifecycle's "QueryChaincodeDefinition" function
		"_lifecycle/QueryChaincodeDefinition": "/Channel/Application/Writers",

		//  ACL policy for _lifecycle's "QueryChaincodeDefinitions" function
		"_lifecycle/QueryChaincodeDefinitions": "/Channel/Application/Writers",

		// ---Lifecycle System Chaincode (lscc) function to policy mapping for access control---//

		//  ACL policy for lscc's "getid" function
		"lscc/ChaincodeExists": "/Channel/Application/Readers",

		//  ACL policy for lscc's "getdepspec" function
		"lscc/GetDeploymentSpec": "/Channel/Application/Readers",

		//  ACL policy for lscc's "getccdata" function
		"lscc/GetChaincodeData": "/Channel/Application/Readers",

		//  ACL Policy for lscc's "getchaincodes" function
		"lscc/GetInstantiatedChaincodes": "/Channel/Application/Readers",

		// ---Query System Chaincode (qscc) function to policy mapping for access control---//

		//  ACL policy for qscc's "GetChainInfo" function
		"qscc/GetChainInfo": "/Channel/Application/Readers",

		//  ACL policy for qscc's "GetBlockByNumber" function
		"qscc/GetBlockByNumber": "/Channel/Application/Readers",

		//  ACL policy for qscc's  "GetBlockByHash" function
		"qscc/GetBlockByHash": "/Channel/Application/Readers",

		//  ACL policy for qscc's "GetTransactionByID" function
		"qscc/GetTransactionByID": "/Channel/Application/Readers",

		//  ACL policy for qscc's "GetBlockByTxID" function
		"qscc/GetBlockByTxID": "/Channel/Application/Readers",

		// ---Configuration System Chaincode (cscc) function to policy mapping for access control---//

		//  ACL policy for cscc's "GetConfigBlock" function
		"cscc/GetConfigBlock": "/Channel/Application/Readers",

		//  ACL policy for cscc's "GetChannelConfig" function
		"cscc/GetChannelConfig": "/Channel/Application/Readers",

		// ---Miscellaneous peer function to policy mapping for access control---//

		//  ACL policy for invoking chaincodes on peer
		"peer/Propose": "/Channel/Application/Writers",

		//  ACL policy for chaincode to chaincode invocation
		"peer/ChaincodeToChaincode": "/Channel/Application/Writers",

		// ---Events resource to policy mapping for access control// // // ---//

		//  ACL policy for sending block events
		"event/Block": "/Channel/Application/Readers",

		//  ACL policy for sending filtered block events
		"event/FilteredBlock": "/Channel/Application/Readers",
	}
}