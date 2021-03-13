package main

import (
        "time"
		"k8s.io/client-go/kubernetes"
        "k8s.io/client-go/rest"
        glog "github.com/sirupsen/logrus"
        "os"
        "strconv"
        "flag"
        "net/http"
        "k8s.io/client-go/tools/clientcmd"
        "github.com/prometheus/client_golang/prometheus"
        "github.com/prometheus/client_golang/prometheus/promauto"
		"github.com/prometheus/client_golang/prometheus/promhttp"
		"context"
		metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
		"regexp"
		//"fmt"
)

var (
        masterURL  string
        kubeconfig string
        inclusterConfig bool
        pollInterval int
)

func init() {
        glog.SetFormatter(&glog.JSONFormatter{})
        glog.SetLevel(glog.InfoLevel)
        getEnvs()
        parseFlags()
}

func main() {

        process()
        http.Handle("/metrics", promhttp.Handler())
        http.ListenAndServe(":2112", nil)
}

func check(err error){
	if err != nil{
		panic("end")
	}
}

func process(){
	autoscale_ready := promauto.NewGaugeVec(prometheus.GaugeOpts{
               Name: "autoscale_ready",
               Help: "autoscale_ready",
        },
		[]string{
			"nodegroup",
		},)

		autoscale_cloudprovidertarget := promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "autoscale_cloudprovidertarget",
			Help: "autoscale_cloudprovidertarget",
		},
		[]string{
			"nodegroup",
		},)

		autoscale_min := promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "autoscale_min",
			Help: "autoscale_min",
		},
		[]string{
			"nodegroup",
		},)
		
		autoscale_max := promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "autoscale_max",
			Help: "autoscale_max",
		},
		[]string{
			"nodegroup",
		},)
		
		autoscale_scaleup_ready := promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "autoscale_scaleup_ready",
			Help: "autoscale_scaleup_ready",
		},
		[]string{
			"nodegroup",
		},)

		autoscale_scaleup_cloudprovidertarget := promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "autoscale_scaleup_cloudprovidertarget",
			Help: "autoscale_scaleup_cloudprovidertarget",
		},
		[]string{
			"nodegroup",
		},)

		autoscale_scaledown_candidates := promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "autoscale_scaledown_candidates",
			Help: "autoscale_scaledown_candidates",
		},
		[]string{
			"nodegroup",
		},)
		

        config, err := getCfg(inclusterConfig, masterURL, kubeconfig)
        if err != nil{
               glog.Errorf("Could not connect to cluster. Error: %v", err)
                 panic(err.Error())
        }

        clientset, err := kubernetes.NewForConfig(config)

        if err != nil {
                glog.Errorf("Error creating client set. Error: %v", err)
               panic(err.Error())
        }

        go func() {
               for {
					configMap, err := clientset.CoreV1().ConfigMaps("kube-system").Get(context.TODO(),"cluster-autoscaler-status", metav1.GetOptions{})
					if err != nil {
						glog.Errorf("Error creating configmap. Error: %v", err)
					   panic(err.Error())
				}
					check(err)
					status, _ := configMap.Data["status"]

					matcher, err := regexp.Compile("Name:\\s*([A-Za-z0-9\\-]+)\n\\s*Health:\\s*\\w*\\s*\\(ready=([0-9])\\s*([A-Za-z0-9\\= ]+)cloudProviderTarget=([0-9])\\s*\\(minSize=([0-9]),\\s*maxSize=([0-9])\\)\\)\n\\s*LastProbeTime:\\s*([0-9\\-]+ [0-9]+:[0-9]+:[0-9]+\\.[0-9]+ \\+[0-9]+ [A-Za-z]+ m=\\+[0-9]+\\.[0-9]+)\n\\s*LastTransitionTime:\\s*([0-9\\-]+ [0-9]+:[0-9]+:[0-9]+\\.[0-9]+ \\+[0-9]+ [A-Za-z]+ m=\\+[0-9]+\\.[0-9]+)\n\\s*ScaleUp:\\s*([A-Za-z]+)\\s*\\(ready=([0-9]+)\\s*cloudProviderTarget=([0-9]+)\\s*\\)\n\\s*LastProbeTime:\\s*([0-9\\-]+ [0-9]+:[0-9]+:[0-9]+\\.[0-9]+ \\+[0-9]+ [A-Za-z]+ m=\\+[0-9]+\\.[0-9]+)\n\\s*LastTransitionTime:\\s*([0-9\\-]+ [0-9]+:[0-9]+:[0-9]+\\.[0-9]+ \\+[0-9]+ [A-Za-z]+ m=\\+[0-9]+\\.[0-9]+)\n\\s*ScaleDown:\\s*([A-Za-z]+)\\s*\\(candidates=([0-9]+)")
					check(err)
					matches := matcher.FindAllStringSubmatch(status, -1)
					
					autoscale_ready.Reset()
					autoscale_cloudprovidertarget.Reset()
					autoscale_min.Reset()
					autoscale_max.Reset()
					autoscale_scaleup_ready.Reset()
					autoscale_scaleup_cloudprovidertarget.Reset()
					autoscale_scaledown_candidates.Reset()

					for _, match := range matches {
							
							autoscale_ready_float, _ := strconv.ParseFloat(string(match[2]),64)
							autoscale_ready.WithLabelValues(string(match[1])).Set(autoscale_ready_float)

							autoscale_cloudprovidertarget_float, _ := strconv.ParseFloat(string(match[4]),64)
							autoscale_cloudprovidertarget.WithLabelValues(string(match[1])).Set(autoscale_cloudprovidertarget_float)

							autoscale_min_float, _ := strconv.ParseFloat(string(match[5]),64)
							autoscale_min.WithLabelValues(string(match[1])).Set(autoscale_min_float)

							autoscale_max_float, _ := strconv.ParseFloat(string(match[6]),64)
							autoscale_max.WithLabelValues(string(match[1])).Set(autoscale_max_float)

							autoscale_scaleup_ready_float, _ := strconv.ParseFloat(string(match[10]),64)
							autoscale_scaleup_ready.WithLabelValues(string(match[1])).Set(autoscale_scaleup_ready_float)

							autoscale_scaleup_cloudprovidertarget_float, _ := strconv.ParseFloat(string(match[11]),64)
							autoscale_scaleup_cloudprovidertarget.WithLabelValues(string(match[1])).Set(autoscale_scaleup_cloudprovidertarget_float)

							autoscale_scaledown_candidates_float, _ := strconv.ParseFloat(string(match[15]),64)
							autoscale_scaledown_candidates.WithLabelValues(string(match[1])).Set(autoscale_scaledown_candidates_float)
					}

                    time.Sleep(time.Second * time.Duration(pollInterval))
               }
        }()

}

func getCfg(inclusterConfig bool, masterURL string, kubeconfig string) (*rest.Config, error) {
        if inclusterConfig {
               return rest.InClusterConfig()
        }

        cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
        if err != nil {
               glog.Errorf("Error building kubeconfig: %s", err.Error())
        }
        return cfg, err
}

func getEnvs() {
        _, err := strconv.Atoi(os.Getenv("POLL_INTERVAL"))
        if err != nil {
               pollInterval = 30
        }else{
               pollInterval, _ = strconv.Atoi(os.Getenv("POLL_INTERVAL"))
        }
        _, inclusterConfig = os.LookupEnv("INCLUSTERCONFIG")
}

func parseFlags() {
        flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
        flag.StringVar(&masterURL, "masterURL", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
        flag.Parse()
}