package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	AnnKeyShowDebug = GroupVersion.Group + "/show-debug"
)

func ShowDebug(m *metav1.ObjectMeta) bool {
	return m.Annotations[AnnKeyShowDebug] != ""
}
