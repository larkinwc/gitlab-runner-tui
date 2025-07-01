package runner

import "time"

type Runner struct {
	ID          string
	Name        string
	Token       string
	Status      string
	Online      bool
	Description string
	TagList     []string
	Locked      bool
	RunsUntagged bool
	Executor    string
	Architecture string
	Platform    string
	LastContact time.Time
	CurrentJob  *Job
}

type Job struct {
	ID         int
	Name       string
	Status     string
	Stage      string
	Project    string
	Pipeline   int
	Started    time.Time
	Finished   time.Time
	Duration   time.Duration
	RunnerName string
	RunnerID   string
	ExitCode   int
	URL        string
}

type Config struct {
	Concurrent    int                    `yaml:"concurrent"`
	CheckInterval int                    `yaml:"check_interval"`
	LogLevel      string                 `yaml:"log_level"`
	LogFormat     string                 `yaml:"log_format"`
	SessionServer SessionServerConfig    `yaml:"session_server,omitempty"`
	Runners       []RunnerConfig         `yaml:"runners"`
}

type SessionServerConfig struct {
	ListenAddress string `yaml:"listen_address"`
	SessionTimeout int   `yaml:"session_timeout"`
}

type RunnerConfig struct {
	Name               string            `yaml:"name"`
	URL                string            `yaml:"url"`
	Token              string            `yaml:"token"`
	Executor           string            `yaml:"executor"`
	Shell              string            `yaml:"shell,omitempty"`
	Builds_dir         string            `yaml:"builds_dir,omitempty"`
	Cache_dir          string            `yaml:"cache_dir,omitempty"`
	Environment        []string          `yaml:"environment,omitempty"`
	Request_concurrency int               `yaml:"request_concurrency,omitempty"`
	Output_limit       int               `yaml:"output_limit,omitempty"`
	Pre_clone_script   string            `yaml:"pre_clone_script,omitempty"`
	Pre_build_script   string            `yaml:"pre_build_script,omitempty"`
	Post_build_script  string            `yaml:"post_build_script,omitempty"`
	Clone_url          string            `yaml:"clone_url,omitempty"`
	Debug_trace_disabled bool            `yaml:"debug_trace_disabled,omitempty"`
	TagList            []string          `yaml:"tag_list,omitempty"`
	RunUntagged        bool              `yaml:"run_untagged,omitempty"`
	Locked             bool              `yaml:"locked,omitempty"`
	Limit              int               `yaml:"limit,omitempty"`
	MaxBuilds          int               `yaml:"max_builds,omitempty"`
	Docker             *DockerConfig     `yaml:"docker,omitempty"`
	Machine            *MachineConfig    `yaml:"machine,omitempty"`
	Kubernetes         *KubernetesConfig `yaml:"kubernetes,omitempty"`
}

type DockerConfig struct {
	Host               string            `yaml:"host,omitempty"`
	Hostname           string            `yaml:"hostname,omitempty"`
	Image              string            `yaml:"image"`
	Runtime            string            `yaml:"runtime,omitempty"`
	Memory             string            `yaml:"memory,omitempty"`
	MemorySwap         string            `yaml:"memory_swap,omitempty"`
	MemoryReservation  string            `yaml:"memory_reservation,omitempty"`
	CpusetCpus         string            `yaml:"cpuset_cpus,omitempty"`
	Cpus               string            `yaml:"cpus,omitempty"`
	DnsSearch          []string          `yaml:"dns_search,omitempty"`
	Dns                []string          `yaml:"dns,omitempty"`
	Privileged         bool              `yaml:"privileged,omitempty"`
	DisableEntrypointOverwrite bool     `yaml:"disable_entrypoint_overwrite,omitempty"`
	Userns             string            `yaml:"userns_mode,omitempty"`
	CapAdd             []string          `yaml:"cap_add,omitempty"`
	CapDrop            []string          `yaml:"cap_drop,omitempty"`
	Oom_kill_disable   bool              `yaml:"oom_kill_disable,omitempty"`
	OomScoreAdjust     int               `yaml:"oom_score_adjust,omitempty"`
	SecurityOpt        []string          `yaml:"security_opt,omitempty"`
	Devices            []string          `yaml:"devices,omitempty"`
	DisableCache       bool              `yaml:"disable_cache,omitempty"`
	Volumes            []string          `yaml:"volumes,omitempty"`
	VolumeDriver       string            `yaml:"volume_driver,omitempty"`
	CacheDir           string            `yaml:"cache_dir,omitempty"`
	ExtraHosts         []string          `yaml:"extra_hosts,omitempty"`
	Isolation          string            `yaml:"isolation,omitempty"`
	NetworkMode        string            `yaml:"network_mode,omitempty"`
	Links              []string          `yaml:"links,omitempty"`
	Services           []string          `yaml:"services,omitempty"`
	WaitForServicesTimeout int          `yaml:"wait_for_services_timeout,omitempty"`
	AllowedImages      []string          `yaml:"allowed_images,omitempty"`
	AllowedServices    []string          `yaml:"allowed_services,omitempty"`
	PullPolicy         []string          `yaml:"pull_policy,omitempty"`
	ShmSize            int64             `yaml:"shm_size,omitempty"`
	Tmpfs              map[string]string `yaml:"tmpfs,omitempty"`
	ServicesTmpfs      map[string]string `yaml:"services_tmpfs,omitempty"`
	SysCtls            map[string]string `yaml:"sysctls,omitempty"`
	HelperImage        string            `yaml:"helper_image,omitempty"`
	HelperImageFlavor  string            `yaml:"helper_image_flavor,omitempty"`
}

type MachineConfig struct {
	IdleCount         int      `yaml:"IdleCount,omitempty"`
	IdleTime          int      `yaml:"IdleTime,omitempty"`
	MaxBuilds         int      `yaml:"MaxBuilds,omitempty"`
	MachineDriver     string   `yaml:"MachineDriver,omitempty"`
	MachineName       string   `yaml:"MachineName,omitempty"`
	MachineOptions    []string `yaml:"MachineOptions,omitempty"`
}

type KubernetesConfig struct {
	Host                          string            `yaml:"host,omitempty"`
	BearerToken                   string            `yaml:"bearer_token,omitempty"`
	BearerTokenOverwriteAllowed   bool              `yaml:"bearer_token_overwrite_allowed,omitempty"`
	Image                         string            `yaml:"image"`
	Namespace                     string            `yaml:"namespace,omitempty"`
	NamespaceOverwriteAllowed     string            `yaml:"namespace_overwrite_allowed,omitempty"`
	Privileged                    bool              `yaml:"privileged,omitempty"`
	RuntimeClassName              string            `yaml:"runtime_class_name,omitempty"`
	AllowPrivilegeEscalation      *bool             `yaml:"allow_privilege_escalation,omitempty"`
	NodeSelector                  map[string]string `yaml:"node_selector,omitempty"`
	NodeTolerations               map[string]string `yaml:"node_tolerations,omitempty"`
	ImagePullSecrets              []string          `yaml:"image_pull_secrets,omitempty"`
	HelperImage                   string            `yaml:"helper_image,omitempty"`
	HelperImageFlavor             string            `yaml:"helper_image_flavor,omitempty"`
	TerminationGracePeriodSeconds int64             `yaml:"terminationGracePeriodSeconds,omitempty"`
	PollInterval                  int               `yaml:"poll_interval,omitempty"`
	PollTimeout                   int               `yaml:"poll_timeout,omitempty"`
	PodLabels                     map[string]string `yaml:"pod_labels,omitempty"`
	ServiceAccount                string            `yaml:"service_account,omitempty"`
	ServiceAccountOverwriteAllowed string           `yaml:"service_account_overwrite_allowed,omitempty"`
	PodAnnotations                map[string]string `yaml:"pod_annotations,omitempty"`
	PodAnnotationsOverwriteAllowed string           `yaml:"pod_annotations_overwrite_allowed,omitempty"`
	PodSecurityContext            map[string]interface{} `yaml:"pod_security_context,omitempty"`
	ResourcesLimits               map[string]string `yaml:"resource_limits,omitempty"`
	ResourcesRequests             map[string]string `yaml:"resource_requests,omitempty"`
	Affinity                      map[string]interface{} `yaml:"affinity,omitempty"`
	Volumes                       map[string]interface{} `yaml:"volumes,omitempty"`
}