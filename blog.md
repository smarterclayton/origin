# Using OpenShift ImageStreams with Kubernetes resources

OpenShift has long offered easy integration between continous integration pipelines that create deployable Docker images and automatic redeployment and rollout with DeploymentConfigs. This makes it easy to define a standard process for continuous deployment that keeps your application always running. As new, higher level constructs like Deployments and StatefulSets have reached maturity in Kubernetes there was no easy way to leverage them and still preserve automatic CI/CD.  

In addition, the image stream concept in OpenShift makes it easy to centralize and manage images that may come from many different locations, but to leverage those images in Kubernetes resources you had to provide the full registry (an internal service IP like 172.17.30.35), the namespace, and the tag of the image, which meant that you didn't get the ease of use that BuildConfigs and DeploymentConfigs offer by allowing direct reference of an image stream tag.

Starting in OpenShift 3.6, we aim to close that gap both by making it as easy to trigger redeployment of Kubernetes Deployments and StatefulSets, and also by allowing Kubernetes resources to easily reference OpenShift image stream tags direcly.

## Triggering image updates on Kubernetes Deployments, StatefulSets, DaemonSets, and more

Image change triggering is a powerful capability. When a new image is pushed to the integrated registry, promoted to latest by a developer's tag, or imported from a remote system as part of a scheduled import, OpenShift automatically identifies any DeploymentConfig or BuildConfig that is "triggered" by that image stream tag, and then updates the image field to point to the latest reference (usually an image "digest", the cryptographic identifier that uniquely describes the contents of the image). This then triggers a new build for a BuildConfig, or a new rollout from a DeploymentConfig.

In OpenShift 3.6 a new alpha feature "annotation triggers" extends this to a number of core Kubernetes resources. A user can set this annotation on their Deployment, and whenever the referenced image stream tag changes the trigger will update the specific referenced container field's "image" to be the value of the image stream tag (a digest URL). The Deployment will then automatically start a new roll-out.

To see this in action, you'll need to be running on an OpenShift 3.6 release - you can try the latest pre-release via `oc cluster up` by downloading the `oc` binary from https://github.com/openshift/origin/releases/v3.6.0-rc.0 and running:

    $ oc cluster up --version=v3.6.0-rc.0

After the cluster has started, let's create a new simple deployment description on disk:

```yaml
kind: Deployment
apiVersion: extensions/v1beta1
metadata:
  name: example
spec:
  template:
    metadata:
      labels:
        app: example
    spec:
      containers:
      - name: web
        image: openshift/deployment-example:v1
        ports:
        - containerPort: 8080
```

Save it as `deployment.yaml`.  The handy example image will show a giant green `v1` when deployed. Deploy it with:

    $ oc create -f deployment.yaml

    # create a service
    $ oc expose deploy/example
    # create a route
    $ oc expose svc/example

then view the newly created route in your browser at `http://example-NAMESPACE.
