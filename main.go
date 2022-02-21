package main

import (
	"context"
	"flag"
	"fmt"
	"net/smtp"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	//HOST üzerinde çalıştırıldığında
	kubeconfig := flag.String("kubeconfig", "kube_config_path", "kubeconfig file path")

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		fmt.Printf("\n[ERROR], %s clientcmd config flag error! :\n ", err.Error())

		//POD olarak çalıştırılacağı zaman
		config, err = rest.InClusterConfig()
		if err != nil {
			fmt.Printf("\n[ERROR], %s InClusterConfig error! :\n ", err.Error())
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("\n[ERROR], %s clientset error! :\n ", err.Error())
	}
	//vars
	ctx := context.Background()
	opts := metav1.ListOptions{}
	var sendermail string
	const receivermail string = "mail_adresi"
	var senderpass string

	namespacelist, err := clientset.CoreV1().Namespaces().List(ctx, opts)
	if err != nil {
		fmt.Printf("\n[ERROR], %s namespace listing error! : ", err.Error())
	}
	if len(namespacelist.Items) == 0 {
		fmt.Printf("\n[ERROR], %s namespace not found! :\n ", err.Error())
	} else {
		// fmt.Println("\nNamespace / Pod List")
		for _, ns := range namespacelist.Items {
			fmt.Println(ns.Name)
			podlist, err := clientset.CoreV1().Pods(ns.Name).List(ctx, opts)
			if err != nil {
				fmt.Printf("\n[ERROR], %s Pod listing error! :\n ", err.Error())
			}
			for _, po := range podlist.Items {
				podCreationTime := po.GetCreationTimestamp()
				age := time.Since(podCreationTime.Time).Round(time.Minute)
				podStatus := po.Status

				var containerRestarts int32
				var containerReady int
				var totalContainers int

				for container := range po.Spec.Containers {
					containerRestarts += podStatus.ContainerStatuses[container].RestartCount
					if podStatus.ContainerStatuses[container].Ready {
						containerReady++
					}
					totalContainers++
				}

				name := po.GetName()
				ready := fmt.Sprintf("%v/%v", containerReady, totalContainers)
				status := fmt.Sprintf("%v", podStatus.Phase)
				restarts := fmt.Sprintf("%v", containerRestarts)
				ageS := age.String()

				data := append([]string{name, ready, status, restarts, ageS})
				fmt.Println(data)

				if len(restarts) > 0 {
					fmt.Println("\nPlease, Enter your email credentials")
					fmt.Println("\nEmail Address: ")
					fmt.Scanln(&sendermail)
					fmt.Println("\nEmail Password: ")
					fmt.Scanln(&senderpass)
					email(sendermail, receivermail, senderpass, data)
				} else {
					email(sendermail, receivermail, senderpass, data)
				}
			}
		}
	}

}

func email(smail string, rmail string, spwd string, dt []string) {
	//Gönderen
	from := *&smail    //test@test.com
	password := *&spwd //passw0rd
	toEmail := *&rmail //test2@test2.com
	to := []string{toEmail}
	//SMTP
	host := "smtp.gmail.com"
	port := "587"
	address := host + ":" + port
	//mesaj
	subj := append([]string{"Subject:", "Cluster", "Pod", "Status", "\n"})
	subject := subj
	body := dt
	// []byte(subject + body)
	//TODO: Mesaj parse etme işlemi düzenlenecek!
	message := []byte(strings.Join(subject, " ") + "\n" + strings.Join(body, " "))
	//authentication
	//funcPlainAuth(identity,username,password,host string) Auth https://pkg.go.dev/net/smtp
	auth := smtp.PlainAuth("", from, password, host)
	//mail gönder
	//func SendMail(addr String, a Auth, from String, to []string, msg []byte)error https://pkg.go.dev/net/smtp
	err := smtp.SendMail(address, auth, from, to, message)
	if err != nil {
		fmt.Printf("\n[ERROR], %s ! ", err.Error())
		return
	} else {
		fmt.Println("Email sent...")
	}

}
