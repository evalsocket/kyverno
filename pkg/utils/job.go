package utils

import (
	corev1 "k8s.io/api/core/v1"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func BuildKyvernoJob(namespace,scope string) *batchv1.Job {
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "random-name",
			Namespace: namespace,
			Labels:    map[string]string{
				"nirmata.io/managed"   : "kyverno",
				"nirmata.io/type" : "job",
			},
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "random-name",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "random-name-job",
							Image: "nirmata/kyvernojob:1.1.8",
							Args: []string{
								"--scan: true",
							},
						},
					},
				},
			},
		},
	}
	return job
}
