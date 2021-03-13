# autoscalescraper

Uses Regex to extract cluster autoscaler nodegroup data from cluster-autoscaler-status configmap in kube-system and make them available via prometheus metrics. Useful when you don't have access to autoscaler metrics i.e. AKS