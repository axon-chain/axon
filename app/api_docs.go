package app

import (
	"bytes"
	"compress/gzip"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"time"

	sdkdocs "github.com/cosmos/cosmos-sdk/client/docs"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/version"
	cosmosproto "github.com/cosmos/gogoproto/proto"
	annotations "github.com/gogo/googleapis/google/api"
	gogoproto "github.com/gogo/protobuf/proto"
	descriptorpb "github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/gorilla/mux"
)

var (
	pathParamPattern = regexp.MustCompile(`\{([^}=]+)(=[^}]+)?\}`)

	apiDocsPageTemplate = template.Must(template.ParseFS(apiDocsTemplateFS, "api_docs.html"))
)

//go:embed api_docs.html
var apiDocsTemplateFS embed.FS

type apiDocsTemplateData struct {
	DataJSON template.JS
}

type apiDocsPayload struct {
	GeneratedAt    string                  `json:"generated_at"`
	Summary        string                  `json:"summary"`
	Runtime        apiDocsRuntime          `json:"runtime"`
	PublicServices []apiDocsServiceEntry   `json:"public_services"`
	LocalServices  []apiDocsLocalService   `json:"local_services"`
	PublicAPI      apiDocsManualSection    `json:"public_api"`
	GeneratedAPI   apiDocsGeneratedSection `json:"generated_api"`
}

type apiDocsRuntime struct {
	DocsURL        string `json:"docs_url"`
	DocsDataURL    string `json:"docs_data_url"`
	PublicAPIBase  string `json:"public_api_base"`
	AgentAPIBase   string `json:"agent_api_base"`
	OpenAPISpecURL string `json:"openapi_spec_url"`
	SwaggerUIURL   string `json:"swagger_ui_url"`
}

type apiDocsServiceEntry struct {
	Name  string `json:"name"`
	Entry string `json:"entry"`
	Notes string `json:"notes"`
}

type apiDocsLocalService struct {
	Name      string `json:"name"`
	Port      string `json:"port"`
	LocalForm string `json:"local_form"`
	Notes     string `json:"notes"`
}

type apiDocsManualSection struct {
	BasePath         string         `json:"base_path"`
	Rules            []apiDocsRule  `json:"rules"`
	ResponseWrappers []apiDocsRule  `json:"response_wrappers"`
	Groups           []apiDocsGroup `json:"groups"`
}

type apiDocsGeneratedSection struct {
	Title   string         `json:"title"`
	Version string         `json:"version"`
	Groups  []apiDocsGroup `json:"groups"`
}

type apiDocsRule struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type apiDocsGroup struct {
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Endpoints   []apiDocsEndpoint `json:"endpoints"`
}

type apiDocsEndpoint struct {
	Method   string   `json:"method"`
	Path     string   `json:"path"`
	Summary  string   `json:"summary"`
	Notes    []string `json:"notes,omitempty"`
	Query    []string `json:"query,omitempty"`
	Response []string `json:"response,omitempty"`
}

type swaggerSpec struct {
	Swagger     string                                 `json:"swagger"`
	Info        swaggerInfo                            `json:"info"`
	Consumes    []string                               `json:"consumes,omitempty"`
	Produces    []string                               `json:"produces,omitempty"`
	Paths       map[string]map[string]swaggerOperation `json:"paths"`
	Definitions map[string]swaggerSchema               `json:"definitions,omitempty"`
}

type swaggerInfo struct {
	Title   string `json:"title"`
	Version string `json:"version"`
}

type swaggerOperation struct {
	Summary     string                     `json:"summary,omitempty"`
	Description string                     `json:"description,omitempty"`
	OperationID string                     `json:"operationId,omitempty"`
	Tags        []string                   `json:"tags,omitempty"`
	Parameters  []swaggerParameter         `json:"parameters,omitempty"`
	Responses   map[string]swaggerResponse `json:"responses,omitempty"`
}

type swaggerParameter struct {
	Name        string         `json:"name,omitempty"`
	In          string         `json:"in,omitempty"`
	Description string         `json:"description,omitempty"`
	Required    bool           `json:"required,omitempty"`
	Type        string         `json:"type,omitempty"`
	Format      string         `json:"format,omitempty"`
	Items       *swaggerSchema `json:"items,omitempty"`
	Schema      *swaggerSchema `json:"schema,omitempty"`
}

type swaggerResponse struct {
	Description string         `json:"description,omitempty"`
	Schema      *swaggerSchema `json:"schema,omitempty"`
}

type swaggerSchema struct {
	Ref                  string                   `json:"$ref,omitempty"`
	Type                 string                   `json:"type,omitempty"`
	Format               string                   `json:"format,omitempty"`
	Description          string                   `json:"description,omitempty"`
	Items                *swaggerSchema           `json:"items,omitempty"`
	Properties           map[string]swaggerSchema `json:"properties,omitempty"`
	Required             []string                 `json:"required,omitempty"`
	Enum                 []string                 `json:"enum,omitempty"`
	AdditionalProperties *swaggerSchema           `json:"additionalProperties,omitempty"`
}

type generatedRoute struct {
	Group     string
	Method    string
	Path      string
	Summary   string
	Operation swaggerOperation
}

type httpBinding struct {
	Method string
	Path   string
	Body   string
}

func registerRuntimeAPIDocs(apiSvr *api.Server) {
	spec, generatedSection, err := buildRuntimeOpenAPISpec()
	if err != nil {
		panic(err)
	}
	registerAPIIndexRoutes(apiSvr.Router)
	registerDocsSiteRoutes(apiSvr.Router, spec, generatedSection)
	if err := registerDocsSwaggerUI(apiSvr.Router, spec); err != nil {
		panic(err)
	}
}

func registerAPIIndexRoutes(router *mux.Router) {
	writeIndex := func(w http.ResponseWriter, payload any) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(payload)
	}

	publicIndex := func(w http.ResponseWriter, r *http.Request) {
		origin := requestOrigin(r)
		writeIndex(w, map[string]any{
			"name":        "Axon Public API",
			"base_path":   publicAPIRoot + "/",
			"docs_url":    origin + "/docs/",
			"openapi_url": origin + "/docs/openapi.json",
			"entrypoints": []string{
				publicAPIRoot + "/chain/info",
				publicAPIRoot + "/chain/status",
				publicAPIRoot + "/blocks/latest",
				publicAPIRoot + "/txs/recent",
				publicAPIRoot + "/validators",
				publicAPIRoot + "/agents",
				publicAPIRoot + "/search",
			},
		})
	}

	agentIndex := func(w http.ResponseWriter, r *http.Request) {
		origin := requestOrigin(r)
		writeIndex(w, map[string]any{
			"name":        "Axon Agent API",
			"base_path":   "/axon/agent/v1/",
			"docs_url":    origin + "/docs/",
			"openapi_url": origin + "/docs/openapi.json",
			"entrypoints": []string{
				"/axon/agent/v1/params",
				"/axon/agent/v1/agents",
				"/axon/agent/v1/agent/{address}",
				"/axon/agent/v1/reputation/{address}",
				"/axon/agent/v1/challenge/current",
			},
		})
	}

	router.HandleFunc(publicAPIRoot, publicIndex).Methods(http.MethodGet)
	router.HandleFunc(publicAPIRoot+"/", publicIndex).Methods(http.MethodGet)
	router.HandleFunc("/axon/agent/v1", agentIndex).Methods(http.MethodGet)
	router.HandleFunc("/axon/agent/v1/", agentIndex).Methods(http.MethodGet)
}

func registerDocsSiteRoutes(router *mux.Router, spec []byte, generatedSection apiDocsGeneratedSection) {
	serveSpec := func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(spec)
	}

	serveDocsData := func(w http.ResponseWriter, r *http.Request) {
		payload := buildAPIDocsPayload(r, generatedSection)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(payload)
	}

	serveDocsPage := func(w http.ResponseWriter, r *http.Request) {
		payload := buildAPIDocsPayload(r, generatedSection)
		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := apiDocsPageTemplate.Execute(w, apiDocsTemplateData{DataJSON: template.JS(string(payloadJSON))}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}

	redirectToRoot := func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs/", http.StatusTemporaryRedirect)
	}

	router.HandleFunc("/docs", redirectToRoot).Methods(http.MethodGet)
	router.HandleFunc("/docs/", serveDocsPage).Methods(http.MethodGet)
	router.HandleFunc("/docs/index.html", serveDocsPage).Methods(http.MethodGet)
	router.HandleFunc("/docs/data.json", serveDocsData).Methods(http.MethodGet)
	router.HandleFunc("/docs/openapi.json", serveSpec).Methods(http.MethodGet)
}

func registerDocsSwaggerUI(router *mux.Router, spec []byte) error {
	root, err := fsSubSwaggerUI()
	if err != nil {
		return err
	}

	staticServer := http.FileServer(root)
	router.PathPrefix("/docs/swagger/").Handler(http.StripPrefix("/docs/swagger/", staticServer))
	router.HandleFunc("/docs/swagger/swagger.yaml", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(spec)
	}).Methods(http.MethodGet)
	return nil
}

func fsSubSwaggerUI() (http.FileSystem, error) {
	root, err := fs.Sub(sdkdocs.SwaggerUI, "swagger-ui")
	if err != nil {
		return nil, err
	}
	return http.FS(root), nil
}

func buildAPIDocsPayload(r *http.Request, generatedSection apiDocsGeneratedSection) apiDocsPayload {
	origin := requestOrigin(r)
	return apiDocsPayload{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Summary:     "Unified runtime API documentation for the mainnet-api entry, including the repository-owned /axon/public/v1 aggregation layer and protobuf-generated Axon agent routes.",
		Runtime: apiDocsRuntime{
			DocsURL:        origin + "/docs/",
			DocsDataURL:    origin + "/docs/data.json",
			PublicAPIBase:  origin + publicAPIRoot + "/",
			AgentAPIBase:   origin + "/axon/agent/v1/",
			OpenAPISpecURL: origin + "/docs/openapi.json",
			SwaggerUIURL:   origin + "/docs/swagger/",
		},
		PublicServices: buildPublicServiceEntries(origin),
		LocalServices:  buildLocalServiceEntries(),
		PublicAPI: apiDocsManualSection{
			BasePath: publicAPIRoot + "/",
			Rules: []apiDocsRule{
				{Label: "Pagination", Value: "limit defaults to 20 and is capped at 100; offset is supported on offset-style endpoints"},
				{Label: "Caching", Value: "selected read-heavy endpoints use short-lived in-process cache; response.source is node or cache"},
				{Label: "Errors", Value: "bad_request, not_found, internal_error"},
				{Label: "Scope", Value: "this layer aggregates current-node data only and does not introduce an external indexer store"},
			},
			ResponseWrappers: []apiDocsRule{
				{Label: "Success", Value: `{"source":"node|cache","generated_at":"RFC3339","data":{...}}`},
				{Label: "Error", Value: `{"code":"bad_request|not_found|internal_error","message":"..."}`},
			},
			Groups: buildManualPublicAPIGroups(),
		},
		GeneratedAPI: generatedSection,
	}
}

func buildPublicServiceEntries(origin string) []apiDocsServiceEntry {
	return []apiDocsServiceEntry{
		{Name: "Unified API Entry", Entry: origin + "/", Notes: "Internally maintained public HTTPS entry for the node REST/API capability set"},
		{Name: "EVM JSON-RPC", Entry: "https://mainnet-rpc.axonchain.ai/", Notes: "Internally maintained public entry mapping to the node EVM JSON-RPC capability on local port 8545"},
		{Name: "Axon Public API", Entry: origin + publicAPIRoot + "/", Notes: "Repository-owned aggregated interface built on current node data"},
		{Name: "Axon Agent API", Entry: origin + "/axon/agent/v1/", Notes: "Generated agent query routes exposed through the same public entry"},
		{Name: "API Docs", Entry: origin + "/docs/", Notes: "Runtime docs site for the same public API entry"},
	}
}

func buildLocalServiceEntries() []apiDocsLocalService {
	return []apiDocsLocalService{
		{Name: "P2P", Port: "26656", LocalForm: "tcp://127.0.0.1:26656", Notes: "Peer connectivity only"},
		{Name: "CometBFT RPC", Port: "26657", LocalForm: "http://127.0.0.1:26657", Notes: "Low-level chain RPC"},
		{Name: "EVM JSON-RPC", Port: "8545", LocalForm: "http://127.0.0.1:8545", Notes: "Wallet and dApp RPC"},
		{Name: "EVM JSON-RPC WebSocket", Port: "8546", LocalForm: "ws://127.0.0.1:8546", Notes: "Local subscription transport"},
		{Name: "Cosmos REST API", Port: "1317", LocalForm: "http://127.0.0.1:1317", Notes: "Standard REST, generated routes, and /axon/public/v1"},
		{Name: "gRPC", Port: "9090", LocalForm: "127.0.0.1:9090", Notes: "Typed service access"},
	}
}

func buildManualPublicAPIGroups() []apiDocsGroup {
	return []apiDocsGroup{
		{
			Name:        "Chain",
			Description: "Static network identity, sync state, fee view, and aggregated chain params.",
			Endpoints: []apiDocsEndpoint{
				{Method: "GET", Path: "/axon/public/v1/chain/info", Summary: "Return chain id, network, version, native token metadata, and bech32 prefixes."},
				{Method: "GET", Path: "/axon/public/v1/chain/status", Summary: "Return latest block height, latest block time, catching_up, peer_count, and node version."},
				{Method: "GET", Path: "/axon/public/v1/chain/health", Summary: "Return lightweight node health, RPC availability, REST availability, and peer count."},
				{Method: "GET", Path: "/axon/public/v1/chain/fees", Summary: "Return current base fee, minimum gas price, and recommended gas price tiers."},
				{Method: "GET", Path: "/axon/public/v1/chain/params", Summary: "Return aggregated staking, slashing, distribution, governance, and agent params."},
			},
		},
		{
			Name:        "Params",
			Description: "Module-specific parameter views without additional aggregation.",
			Endpoints: []apiDocsEndpoint{
				{Method: "GET", Path: "/axon/public/v1/params/staking", Summary: "Return staking params."},
				{Method: "GET", Path: "/axon/public/v1/params/slashing", Summary: "Return slashing params."},
				{Method: "GET", Path: "/axon/public/v1/params/distribution", Summary: "Return distribution params."},
				{Method: "GET", Path: "/axon/public/v1/params/agent", Summary: "Return agent module params."},
				{Method: "GET", Path: "/axon/public/v1/gov/params", Summary: "Return governance deposit, voting, and tally params."},
			},
		},
		{
			Name:        "Blocks",
			Description: "Current-node block lookup, range listing, block txs, validator set, and proposer view.",
			Endpoints: []apiDocsEndpoint{
				{Method: "GET", Path: "/axon/public/v1/blocks/latest", Summary: "Return the latest block summary."},
				{Method: "GET", Path: "/axon/public/v1/blocks/{identifier}", Summary: "Lookup a block by height or hex hash."},
				{Method: "GET", Path: "/axon/public/v1/blocks", Summary: "Return a descending block window.", Query: []string{"from", "to", "limit"}},
				{Method: "GET", Path: "/axon/public/v1/blocks/{height}/txs", Summary: "Return block metadata and tx service results for one height.", Query: []string{"limit", "offset"}},
				{Method: "GET", Path: "/axon/public/v1/blocks/{height}/validators", Summary: "Return validator set at a target height.", Query: []string{"limit", "offset"}},
				{Method: "GET", Path: "/axon/public/v1/blocks/{height}/proposer", Summary: "Return proposer consensus address and matching validator-set entry when available."},
			},
		},
		{
			Name:        "Transactions",
			Description: "Current-node tx lookup, event query, simulate, and broadcast helpers.",
			Endpoints: []apiDocsEndpoint{
				{Method: "GET", Path: "/axon/public/v1/txs/recent", Summary: "Return recent indexed transactions.", Query: []string{"limit"}},
				{Method: "GET", Path: "/axon/public/v1/txs/search", Summary: "Run a direct tx event query string.", Query: []string{"q", "limit", "offset"}},
				{Method: "GET", Path: "/axon/public/v1/txs", Summary: "Build a tx search query from common filters.", Query: []string{"sender", "recipient", "type", "status", "from_height", "to_height", "limit", "offset"}},
				{Method: "GET", Path: "/axon/public/v1/txs/{hash}", Summary: "Return a stable tx body summary and tx execution summary."},
				{Method: "GET", Path: "/axon/public/v1/txs/{hash}/events", Summary: "Return tx events only."},
				{Method: "GET", Path: "/axon/public/v1/txs/{hash}/raw", Summary: "Return base64-encoded protobuf tx bytes."},
				{Method: "POST", Path: "/axon/public/v1/txs/simulate", Summary: "Simulate signed tx bytes.", Notes: []string{`body: {"tx_bytes":"<base64-or-hex>"}`}},
				{Method: "POST", Path: "/axon/public/v1/txs/broadcast", Summary: "Broadcast signed tx bytes.", Notes: []string{`body: {"tx_bytes":"<base64-or-hex>","mode":"sync|async|block"}`}},
			},
		},
		{
			Name:        "Accounts",
			Description: "Stable base-account summary, balances, tx lookups, and rewards.",
			Endpoints: []apiDocsEndpoint{
				{Method: "GET", Path: "/axon/public/v1/accounts/{address}", Summary: "Return a stable base-account summary."},
				{Method: "GET", Path: "/axon/public/v1/accounts/{address}/balances", Summary: "Return all balances.", Query: []string{"limit", "offset"}},
				{Method: "GET", Path: "/axon/public/v1/accounts/{address}/spendable", Summary: "Return spendable balances.", Query: []string{"limit", "offset"}},
				{Method: "GET", Path: "/axon/public/v1/accounts/{address}/sequence", Summary: "Return account_number and sequence."},
				{Method: "GET", Path: "/axon/public/v1/accounts/{address}/txs", Summary: "Return merged account tx activity from current-node tx index.", Query: []string{"limit"}},
				{Method: "GET", Path: "/axon/public/v1/accounts/{address}/transfers", Summary: "Return merged transfer-related tx activity.", Query: []string{"limit"}},
				{Method: "GET", Path: "/axon/public/v1/accounts/{address}/rewards", Summary: "Return distribution rewards summary."},
			},
		},
		{
			Name:        "Validators",
			Description: "Validator list, ranking, detail, signing info, rewards, and Axon agent mapping.",
			Endpoints: []apiDocsEndpoint{
				{Method: "GET", Path: "/axon/public/v1/validators", Summary: "Return validators.", Query: []string{"status", "limit", "offset"}},
				{Method: "GET", Path: "/axon/public/v1/validators/top", Summary: "Return bonded validators sorted by staked tokens.", Query: []string{"limit"}},
				{Method: "GET", Path: "/axon/public/v1/validators/{valoper}", Summary: "Return validator summary, consensus pubkey summary, commission, self-delegation, signing info, and mapped agent if present."},
				{Method: "GET", Path: "/axon/public/v1/validators/{valoper}/status", Summary: "Return validator status and unbonding time."},
				{Method: "GET", Path: "/axon/public/v1/validators/{valoper}/delegations", Summary: "Return validator delegations.", Query: []string{"limit", "offset"}},
				{Method: "GET", Path: "/axon/public/v1/validators/{valoper}/unbondings", Summary: "Return validator unbondings.", Query: []string{"limit", "offset"}},
				{Method: "GET", Path: "/axon/public/v1/validators/{valoper}/redelegations", Summary: "Return validator redelegations aggregated from source and destination lookups.", Query: []string{"limit", "offset"}},
				{Method: "GET", Path: "/axon/public/v1/validators/{valoper}/commission", Summary: "Return validator commission."},
				{Method: "GET", Path: "/axon/public/v1/validators/{valoper}/rewards", Summary: "Return validator outstanding rewards."},
				{Method: "GET", Path: "/axon/public/v1/validators/{valoper}/slashes", Summary: "Return validator slash events.", Query: []string{"from_height", "to_height", "limit", "offset"}},
				{Method: "GET", Path: "/axon/public/v1/validators/{valoper}/signing-info", Summary: "Return slashing signing info and signed window values."},
				{Method: "GET", Path: "/axon/public/v1/validators/{valoper}/self-delegation", Summary: "Return the self-delegation by mapping valoper to the same account bytes."},
				{Method: "GET", Path: "/axon/public/v1/validators/{valoper}/agent", Summary: "Return the mapped Axon agent if present."},
				{Method: "GET", Path: "/axon/public/v1/validators/{valoper}/uptime", Summary: "Return a lightweight uptime estimate derived from slashing signing info."},
			},
		},
		{
			Name:        "Delegators",
			Description: "Delegator-centric staking, reward, withdraw-address, and validator views.",
			Endpoints: []apiDocsEndpoint{
				{Method: "GET", Path: "/axon/public/v1/delegators/{address}/delegations", Summary: "Return delegations.", Query: []string{"limit", "offset"}},
				{Method: "GET", Path: "/axon/public/v1/delegators/{address}/unbondings", Summary: "Return unbonding delegations.", Query: []string{"limit", "offset"}},
				{Method: "GET", Path: "/axon/public/v1/delegators/{address}/redelegations", Summary: "Return redelegations.", Query: []string{"limit", "offset"}},
				{Method: "GET", Path: "/axon/public/v1/delegators/{address}/rewards", Summary: "Return delegator rewards."},
				{Method: "GET", Path: "/axon/public/v1/delegators/{address}/withdraw-address", Summary: "Return withdraw address."},
				{Method: "GET", Path: "/axon/public/v1/delegators/{address}/validators", Summary: "Return validators related to the delegator.", Query: []string{"limit", "offset"}},
			},
		},
		{
			Name:        "Governance",
			Description: "Proposal list, proposal detail, votes, tally, and governance params.",
			Endpoints: []apiDocsEndpoint{
				{Method: "GET", Path: "/axon/public/v1/gov/proposals", Summary: "Return proposals.", Query: []string{"status", "voter", "depositor", "limit", "offset"}},
				{Method: "GET", Path: "/axon/public/v1/gov/proposals/{id}", Summary: "Return one proposal.", Notes: []string{"missing proposal returns 404 not_found"}},
				{Method: "GET", Path: "/axon/public/v1/gov/proposals/{id}/votes", Summary: "Return proposal votes.", Query: []string{"limit", "offset"}, Notes: []string{"missing proposal returns 404 not_found"}},
				{Method: "GET", Path: "/axon/public/v1/gov/proposals/{id}/tally", Summary: "Return proposal tally.", Notes: []string{"missing proposal returns 404 not_found"}},
				{Method: "GET", Path: "/axon/public/v1/gov/params", Summary: "Return governance params."},
			},
		},
		{
			Name:        "Explorer",
			Description: "Compact dashboard-oriented views built directly from current node state.",
			Endpoints: []apiDocsEndpoint{
				{Method: "GET", Path: "/axon/public/v1/explorer/overview", Summary: "Return chain, validator, and agent overview."},
				{Method: "GET", Path: "/axon/public/v1/explorer/stats", Summary: "Return recent block span, tx count, peer count, and current store metadata."},
				{Method: "GET", Path: "/axon/public/v1/explorer/validators/top", Summary: "Alias of validator ranking.", Query: []string{"limit"}},
				{Method: "GET", Path: "/axon/public/v1/explorer/blocks/recent", Summary: "Return recent blocks.", Query: []string{"limit"}},
				{Method: "GET", Path: "/axon/public/v1/explorer/txs/recent", Summary: "Return recent indexed txs.", Query: []string{"limit"}},
			},
		},
		{
			Name:        "Agents",
			Description: "Axon-specific agent discovery, heartbeat, reputation, and challenge views.",
			Endpoints: []apiDocsEndpoint{
				{Method: "GET", Path: "/axon/public/v1/agents", Summary: "Return agent list.", Query: []string{"status", "offset", "limit"}},
				{Method: "GET", Path: "/axon/public/v1/agents/{address}", Summary: "Return one agent summary."},
				{Method: "GET", Path: "/axon/public/v1/agents/{address}/heartbeat", Summary: "Return heartbeat-derived status and remaining blocks until offline."},
				{Method: "GET", Path: "/axon/public/v1/agents/{address}/reputation", Summary: "Return agent reputation view."},
				{Method: "GET", Path: "/axon/public/v1/agents/{address}/stake", Summary: "Return agent stake view."},
				{Method: "GET", Path: "/axon/public/v1/agents/online-validators", Summary: "Return agents currently considered online."},
				{Method: "GET", Path: "/axon/public/v1/agents/challenge/current", Summary: "Return current on-chain challenge if present."},
			},
		},
		{
			Name:        "Search",
			Description: "Lightweight search over current-node-supported entities.",
			Endpoints: []apiDocsEndpoint{
				{Method: "GET", Path: "/axon/public/v1/search", Summary: "Resolve tx hash, account address, validator operator address, agent id, or validator moniker.", Query: []string{"q"}},
			},
		},
	}
}

func buildRuntimeOpenAPISpec() ([]byte, apiDocsGeneratedSection, error) {
	definitions := buildCommonDefinitions()

	manualPaths := buildManualOpenAPIPaths(definitions)
	generatedPaths, generatedSection, err := buildGeneratedAgentOpenAPI(definitions)
	if err != nil {
		return nil, apiDocsGeneratedSection{}, err
	}

	for path, methods := range generatedPaths {
		existing, ok := manualPaths[path]
		if !ok {
			existing = make(map[string]swaggerOperation)
			manualPaths[path] = existing
		}
		for method, op := range methods {
			existing[method] = op
		}
	}

	spec := swaggerSpec{
		Swagger:     "2.0",
		Info:        swaggerInfo{Title: "Axon Runtime API", Version: version.Version},
		Consumes:    []string{"application/json"},
		Produces:    []string{"application/json"},
		Paths:       manualPaths,
		Definitions: definitions,
	}

	payload, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		return nil, apiDocsGeneratedSection{}, err
	}
	return payload, generatedSection, nil
}

func buildCommonDefinitions() map[string]swaggerSchema {
	return map[string]swaggerSchema{
		"PublicAPIEnvelope": {
			Type: "object",
			Properties: map[string]swaggerSchema{
				"source":       {Type: "string"},
				"generated_at": {Type: "string", Format: "date-time"},
				"data":         {Type: "object"},
			},
			Required: []string{"source", "generated_at", "data"},
		},
		"ErrorResponse": {
			Type: "object",
			Properties: map[string]swaggerSchema{
				"code":    {Type: "string"},
				"message": {Type: "string"},
			},
			Required: []string{"code", "message"},
		},
		"TxSimulateRequest": {
			Type: "object",
			Properties: map[string]swaggerSchema{
				"tx_bytes": {Type: "string"},
			},
			Required: []string{"tx_bytes"},
		},
		"TxBroadcastRequest": {
			Type: "object",
			Properties: map[string]swaggerSchema{
				"tx_bytes": {Type: "string"},
				"mode":     {Type: "string", Enum: []string{"sync", "async", "block"}},
			},
			Required: []string{"tx_bytes"},
		},
	}
}

func buildManualOpenAPIPaths(definitions map[string]swaggerSchema) map[string]map[string]swaggerOperation {
	paths := make(map[string]map[string]swaggerOperation)
	for _, group := range buildManualPublicAPIGroups() {
		tag := group.Name
		for _, endpoint := range group.Endpoints {
			method := strings.ToLower(endpoint.Method)
			if paths[endpoint.Path] == nil {
				paths[endpoint.Path] = make(map[string]swaggerOperation)
			}
			op := swaggerOperation{
				Summary:     endpoint.Summary,
				Description: strings.Join(endpoint.Notes, " "),
				OperationID: buildOperationID("public", endpoint.Method, endpoint.Path),
				Tags:        []string{tag},
				Parameters:  buildManualParameters(endpoint),
				Responses: map[string]swaggerResponse{
					"200": {
						Description: "OK",
						Schema:      &swaggerSchema{Ref: "#/definitions/PublicAPIEnvelope"},
					},
					"default": {
						Description: "Error",
						Schema:      &swaggerSchema{Ref: "#/definitions/ErrorResponse"},
					},
				},
			}
			if endpoint.Method == http.MethodPost {
				switch endpoint.Path {
				case publicAPIRoot + "/txs/simulate":
					op.Parameters = append(op.Parameters, swaggerParameter{
						Name:     "body",
						In:       "body",
						Required: true,
						Schema:   &swaggerSchema{Ref: "#/definitions/TxSimulateRequest"},
					})
				case publicAPIRoot + "/txs/broadcast":
					op.Parameters = append(op.Parameters, swaggerParameter{
						Name:     "body",
						In:       "body",
						Required: true,
						Schema:   &swaggerSchema{Ref: "#/definitions/TxBroadcastRequest"},
					})
				}
			}
			paths[endpoint.Path][method] = op
		}
	}

	return paths
}

func buildGeneratedAgentOpenAPI(definitions map[string]swaggerSchema) (map[string]map[string]swaggerOperation, apiDocsGeneratedSection, error) {
	fd, err := loadRegisteredFileDescriptor("axon/agent/v1/query.proto")
	if err != nil {
		return nil, apiDocsGeneratedSection{}, err
	}

	var svc *descriptorpb.ServiceDescriptorProto
	for _, candidate := range fd.GetService() {
		if candidate.GetName() == "Query" {
			svc = candidate
			break
		}
	}
	if svc == nil {
		return nil, apiDocsGeneratedSection{}, fmt.Errorf("query service not found in descriptor")
	}

	groupName := "Axon Agent Query"
	paths := make(map[string]map[string]swaggerOperation)
	routes := make([]generatedRoute, 0)

	for _, method := range svc.GetMethod() {
		rule, err := extractHTTPRule(method)
		if err != nil {
			return nil, apiDocsGeneratedSection{}, err
		}

		bindings := flattenHTTPRule(rule)
		if len(bindings) == 0 {
			bindings = fallbackGeneratedHTTPBindings(method.GetName())
		}
		if len(bindings) == 0 {
			continue
		}

		for bindingIndex, binding := range bindings {
			inputType := strings.TrimPrefix(method.GetInputType(), ".")
			outputType := strings.TrimPrefix(method.GetOutputType(), ".")
			parameters := buildGeneratedParameters(inputType, binding, definitions)
			outputRef := ensureMessageSchema(definitions, outputType)

			op := swaggerOperation{
				Summary:     buildGeneratedSummary(method.GetName(), binding.Path),
				Description: "Generated at startup from protobuf HTTP annotations.",
				OperationID: buildOperationID(method.GetName(), binding.Method, fmt.Sprintf("%s_%d", binding.Path, bindingIndex)),
				Tags:        []string{groupName},
				Parameters:  parameters,
				Responses: map[string]swaggerResponse{
					"200": {
						Description: "OK",
						Schema:      &swaggerSchema{Ref: outputRef},
					},
					"default": {
						Description: "Error",
						Schema:      &swaggerSchema{Ref: "#/definitions/ErrorResponse"},
					},
				},
			}

			if paths[binding.Path] == nil {
				paths[binding.Path] = make(map[string]swaggerOperation)
			}
			paths[binding.Path][strings.ToLower(binding.Method)] = op
			routes = append(routes, generatedRoute{
				Group:     groupName,
				Method:    binding.Method,
				Path:      binding.Path,
				Summary:   op.Summary,
				Operation: op,
			})
		}
	}

	sort.Slice(routes, func(i, j int) bool {
		if routes[i].Path == routes[j].Path {
			return routes[i].Method < routes[j].Method
		}
		return routes[i].Path < routes[j].Path
	})

	endpoints := make([]apiDocsEndpoint, 0, len(routes))
	for _, route := range routes {
		endpoints = append(endpoints, apiDocsEndpoint{
			Method:   route.Method,
			Path:     route.Path,
			Summary:  route.Summary,
			Query:    formatSwaggerParameters(route.Operation.Parameters),
			Response: formatSwaggerResponses(route.Operation.Responses),
		})
	}

	return paths, apiDocsGeneratedSection{
		Title:   "Generated Axon Agent API",
		Version: version.Version,
		Groups: []apiDocsGroup{
			{
				Name:        groupName,
				Description: "Generated at service startup from protobuf descriptors and HTTP annotations.",
				Endpoints:   endpoints,
			},
		},
	}, nil
}

func buildGeneratedParameters(inputType string, binding httpBinding, definitions map[string]swaggerSchema) []swaggerParameter {
	pathParams := extractPathParams(binding.Path)
	pathSet := make(map[string]struct{}, len(pathParams))
	parameters := make([]swaggerParameter, 0, len(pathParams)+4)

	for _, name := range pathParams {
		pathSet[name] = struct{}{}
		parameters = append(parameters, swaggerParameter{
			Name:     name,
			In:       "path",
			Required: true,
			Type:     "string",
		})
	}

	if binding.Body != "" {
		parameters = append(parameters, swaggerParameter{
			Name:     "body",
			In:       "body",
			Required: true,
			Schema:   &swaggerSchema{Ref: ensureMessageSchema(definitions, inputType)},
		})
		return parameters
	}

	for _, field := range listMessageFields(inputType) {
		if _, ok := pathSet[field.Name]; ok {
			continue
		}
		parameters = append(parameters, field)
	}

	return parameters
}

func listMessageFields(fullName string) []swaggerParameter {
	msgType := cosmosproto.MessageType(fullName)
	if msgType == nil {
		return nil
	}

	if msgType.Kind() == reflect.Pointer {
		msgType = msgType.Elem()
	}

	fields := make([]swaggerParameter, 0, msgType.NumField())
	for i := 0; i < msgType.NumField(); i++ {
		field := msgType.Field(i)
		if strings.HasPrefix(field.Name, "XXX_") || !field.IsExported() {
			continue
		}

		jsonName := jsonFieldName(field)
		if jsonName == "" {
			continue
		}

		schema := schemaFromType(field.Type)
		param := swaggerParameter{
			Name: jsonName,
			In:   "query",
		}
		if schema.Type != "" {
			param.Type = schema.Type
			param.Format = schema.Format
			param.Items = schema.Items
		} else {
			param.Type = "string"
		}
		fields = append(fields, param)
	}

	sort.Slice(fields, func(i, j int) bool { return fields[i].Name < fields[j].Name })
	return fields
}

func loadRegisteredFileDescriptor(name string) (*descriptorpb.FileDescriptorProto, error) {
	compressed := cosmosproto.FileDescriptor(name)
	if len(compressed) == 0 {
		return nil, fmt.Errorf("registered descriptor not found: %s", name)
	}

	reader, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	raw, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	var fd descriptorpb.FileDescriptorProto
	if err := cosmosproto.Unmarshal(raw, &fd); err != nil {
		return nil, err
	}
	return &fd, nil
}

func extractHTTPRule(method *descriptorpb.MethodDescriptorProto) (*annotations.HttpRule, error) {
	if method.GetOptions() == nil || !gogoproto.HasExtension(method.GetOptions(), annotations.E_Http) {
		return nil, nil
	}

	ext, err := gogoproto.GetExtension(method.GetOptions(), annotations.E_Http)
	if err != nil {
		return nil, err
	}

	rule, ok := ext.(*annotations.HttpRule)
	if !ok {
		return nil, fmt.Errorf("unexpected http rule type for method %s", method.GetName())
	}
	return rule, nil
}

func flattenHTTPRule(rule *annotations.HttpRule) []httpBinding {
	if rule == nil {
		return nil
	}

	binding := httpBindingFromRule(rule)
	bindings := make([]httpBinding, 0, 1+len(rule.GetAdditionalBindings()))
	if binding.Method != "" && binding.Path != "" {
		bindings = append(bindings, binding)
	}
	for _, extra := range rule.GetAdditionalBindings() {
		bindings = append(bindings, flattenHTTPRule(extra)...)
	}
	return bindings
}

func fallbackGeneratedHTTPBindings(methodName string) []httpBinding {
	switch methodName {
	case "Params":
		return []httpBinding{{Method: http.MethodGet, Path: "/axon/agent/v1/params"}}
	case "Agent":
		return []httpBinding{{Method: http.MethodGet, Path: "/axon/agent/v1/agent/{address}"}}
	case "Agents":
		return []httpBinding{{Method: http.MethodGet, Path: "/axon/agent/v1/agents"}}
	case "Reputation":
		return []httpBinding{{Method: http.MethodGet, Path: "/axon/agent/v1/reputation/{address}"}}
	case "CurrentChallenge":
		return []httpBinding{{Method: http.MethodGet, Path: "/axon/agent/v1/challenge/current"}}
	default:
		return nil
	}
}

func httpBindingFromRule(rule *annotations.HttpRule) httpBinding {
	if rule == nil {
		return httpBinding{}
	}

	switch pattern := rule.GetPattern().(type) {
	case *annotations.HttpRule_Get:
		return httpBinding{Method: http.MethodGet, Path: pattern.Get, Body: rule.GetBody()}
	case *annotations.HttpRule_Post:
		return httpBinding{Method: http.MethodPost, Path: pattern.Post, Body: rule.GetBody()}
	case *annotations.HttpRule_Put:
		return httpBinding{Method: http.MethodPut, Path: pattern.Put, Body: rule.GetBody()}
	case *annotations.HttpRule_Delete:
		return httpBinding{Method: http.MethodDelete, Path: pattern.Delete, Body: rule.GetBody()}
	case *annotations.HttpRule_Patch:
		return httpBinding{Method: http.MethodPatch, Path: pattern.Patch, Body: rule.GetBody()}
	case *annotations.HttpRule_Custom:
		return httpBinding{Method: strings.ToUpper(pattern.Custom.GetKind()), Path: pattern.Custom.GetPath(), Body: rule.GetBody()}
	default:
		return httpBinding{}
	}
}

func extractPathParams(path string) []string {
	matches := pathParamPattern.FindAllStringSubmatch(path, -1)
	items := make([]string, 0, len(matches))
	seen := make(map[string]struct{}, len(matches))
	for _, match := range matches {
		name := match[1]
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		items = append(items, name)
	}
	return items
}

func buildGeneratedSummary(methodName string, path string) string {
	switch methodName {
	case "Params":
		return "Return agent module params."
	case "Agent":
		return "Return one agent by address."
	case "Agents":
		return "Return the full agent list."
	case "Reputation":
		return "Return one agent reputation view."
	case "CurrentChallenge":
		return "Return the current on-chain challenge."
	default:
		return fmt.Sprintf("Generated route for %s on %s.", methodName, path)
	}
}

func ensureMessageSchema(definitions map[string]swaggerSchema, fullName string) string {
	defName := schemaDefinitionName(fullName)
	ref := "#/definitions/" + defName
	if _, ok := definitions[defName]; ok {
		return ref
	}

	msgType := cosmosproto.MessageType(fullName)
	if msgType == nil {
		definitions[defName] = swaggerSchema{Type: "object"}
		return ref
	}

	if msgType.Kind() == reflect.Pointer {
		msgType = msgType.Elem()
	}

	properties := make(map[string]swaggerSchema)
	required := make([]string, 0)

	for i := 0; i < msgType.NumField(); i++ {
		field := msgType.Field(i)
		if strings.HasPrefix(field.Name, "XXX_") || !field.IsExported() {
			continue
		}
		jsonName := jsonFieldName(field)
		if jsonName == "" {
			continue
		}

		properties[jsonName] = schemaFromGoType(field.Type, definitions)
		if !strings.Contains(field.Tag.Get("json"), "omitempty") {
			required = append(required, jsonName)
		}
	}

	sort.Strings(required)
	definitions[defName] = swaggerSchema{
		Type:       "object",
		Properties: properties,
		Required:   required,
	}
	return ref
}

func schemaFromGoType(t reflect.Type, definitions map[string]swaggerSchema) swaggerSchema {
	if t.Kind() == reflect.Pointer {
		return schemaFromGoType(t.Elem(), definitions)
	}

	switch {
	case isBytesType(t):
		return swaggerSchema{Type: "string", Format: "byte"}
	case isCoinType(t):
		return swaggerSchema{
			Type: "object",
			Properties: map[string]swaggerSchema{
				"denom":  {Type: "string"},
				"amount": {Type: "string"},
			},
			Required: []string{"denom", "amount"},
		}
	case isJSONScalarWrapper(t):
		return swaggerSchema{Type: "string"}
	case isEnumType(t):
		return swaggerSchema{Type: "string"}
	}

	switch t.Kind() {
	case reflect.Bool:
		return swaggerSchema{Type: "boolean"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		return swaggerSchema{Type: "integer", Format: "int32"}
	case reflect.Int64:
		return swaggerSchema{Type: "string", Format: "int64"}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32:
		return swaggerSchema{Type: "integer", Format: "int32"}
	case reflect.Uint64:
		return swaggerSchema{Type: "string", Format: "uint64"}
	case reflect.Float32:
		return swaggerSchema{Type: "number", Format: "float"}
	case reflect.Float64:
		return swaggerSchema{Type: "number", Format: "double"}
	case reflect.String:
		return swaggerSchema{Type: "string"}
	case reflect.Slice, reflect.Array:
		return swaggerSchema{
			Type:  "array",
			Items: ptrSchema(schemaFromGoType(t.Elem(), definitions)),
		}
	case reflect.Map:
		return swaggerSchema{
			Type:                 "object",
			AdditionalProperties: ptrSchema(schemaFromGoType(t.Elem(), definitions)),
		}
	case reflect.Struct:
		if fullName := registeredMessageName(t); fullName != "" {
			return swaggerSchema{Ref: ensureMessageSchema(definitions, fullName)}
		}
		return swaggerSchema{Type: "object"}
	default:
		return swaggerSchema{Type: "string"}
	}
}

func schemaFromType(t reflect.Type) swaggerSchema {
	return schemaFromGoType(t, map[string]swaggerSchema{})
}

func isBytesType(t reflect.Type) bool {
	return t.Kind() == reflect.Slice && t.Elem().Kind() == reflect.Uint8
}

func isCoinType(t reflect.Type) bool {
	return t.PkgPath() == "github.com/cosmos/cosmos-sdk/types" && t.Name() == "Coin"
}

func isJSONScalarWrapper(t reflect.Type) bool {
	if t.PkgPath() == "time" && t.Name() == "Time" {
		return true
	}

	marshalJSON := reflect.TypeOf((*json.Marshaler)(nil)).Elem()
	return t.Implements(marshalJSON) || reflect.PointerTo(t).Implements(marshalJSON)
}

func isEnumType(t reflect.Type) bool {
	return t.Kind() == reflect.Int32 && t.Name() != ""
}

func registeredMessageName(t reflect.Type) string {
	msgType := t
	if msgType.Kind() != reflect.Pointer {
		msgType = reflect.PointerTo(msgType)
	}

	value := reflect.New(msgType.Elem()).Interface()
	msg, ok := value.(cosmosproto.Message)
	if !ok {
		return ""
	}
	return cosmosproto.MessageName(msg)
}

func schemaDefinitionName(fullName string) string {
	return strings.ReplaceAll(fullName, ".", "_")
}

func buildManualParameters(endpoint apiDocsEndpoint) []swaggerParameter {
	parameters := make([]swaggerParameter, 0, len(endpoint.Query)+2)
	for _, name := range extractPathParams(endpoint.Path) {
		parameters = append(parameters, swaggerParameter{
			Name:     name,
			In:       "path",
			Required: true,
			Type:     "string",
		})
	}
	for _, name := range endpoint.Query {
		schema := manualQuerySchema(name)
		parameters = append(parameters, swaggerParameter{
			Name:   name,
			In:     "query",
			Type:   schema.Type,
			Format: schema.Format,
			Items:  schema.Items,
		})
	}
	return parameters
}

func manualQuerySchema(name string) swaggerSchema {
	switch name {
	case "limit", "offset", "from", "to", "from_height", "to_height", "id":
		return swaggerSchema{Type: "integer", Format: "int64"}
	default:
		return swaggerSchema{Type: "string"}
	}
}

func buildOperationID(prefix, method, path string) string {
	clean := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return r
		case r >= 'A' && r <= 'Z':
			return r
		case r >= '0' && r <= '9':
			return r
		default:
			return '_'
		}
	}, prefix+"_"+method+"_"+path)
	return clean
}

func formatSwaggerParameters(parameters []swaggerParameter) []string {
	items := make([]string, 0, len(parameters))
	for _, parameter := range parameters {
		label := fmt.Sprintf("%s(%s", parameter.Name, parameter.In)
		switch {
		case parameter.Type != "":
			label += ", " + parameter.Type
		case parameter.Schema != nil && parameter.Schema.Ref != "":
			label += ", " + strings.TrimPrefix(parameter.Schema.Ref, "#/definitions/")
		case parameter.Schema != nil && parameter.Schema.Type != "":
			label += ", " + parameter.Schema.Type
		}
		if parameter.Required {
			label += ", required"
		}
		label += ")"
		items = append(items, label)
	}
	return items
}

func formatSwaggerResponses(responses map[string]swaggerResponse) []string {
	codes := make([]string, 0, len(responses))
	for code := range responses {
		codes = append(codes, code)
	}
	sort.Strings(codes)

	items := make([]string, 0, len(codes))
	for _, code := range codes {
		response := responses[code]
		summary := code
		switch {
		case response.Schema != nil && response.Schema.Ref != "":
			summary += " -> " + strings.TrimPrefix(response.Schema.Ref, "#/definitions/")
		case response.Schema != nil && response.Schema.Type != "":
			summary += " -> " + response.Schema.Type
		case response.Description != "":
			summary += " -> " + response.Description
		}
		items = append(items, summary)
	}
	return items
}

func jsonFieldName(field reflect.StructField) string {
	tag := field.Tag.Get("json")
	if tag == "" || tag == "-" {
		return ""
	}
	name := strings.Split(tag, ",")[0]
	if name == "" {
		return ""
	}
	return name
}

func ptrSchema(schema swaggerSchema) *swaggerSchema {
	return &schema
}

func requestOrigin(r *http.Request) string {
	scheme := "http"
	if forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")); forwarded != "" {
		scheme = strings.Split(forwarded, ",")[0]
	} else if r.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + r.Host
}
