# Appsody operator test plan
- Variations:
    - Operator already installed in the ns
    - Operator already watching the ns
    - App already deployed in the ns

- Test number: 236a
    - Title: New deploy default
    - Issue: 236,238, 241
    - Prereq:
        - start from new project directory
        - `appsody init <stack>`
    - Steps:
        1. `appsody deploy`
            - operator should be installed in namespace default watching namespace default
            - app should be deployed to namespace default
        2. `appsody operator install`
            - this should fail with an error saying "operator exists in namespace default watching namespace default"
        3. `appsody operator uninstall`
            - this should fail with an error saying "unable to uninstall operator with running apps. use --force option to force the uninstall and remove apps"
        4. `appsody operator uninstall --force`
            - operator should be removed from namespace default
            - app should be removed from namespace default

- Test number: 236b
    - Title: New deploy custom namespace
    - Issue: 236, 238, 241
    - Prereq:
        - start from new project directory
        - `appsody init <stack>`
        - `kubectl create ns test`
    - Steps:
        1. `appsody deploy --namespace test`
            - operator should be installed in namespace test watching namespace test
            - app should be deployed to namespace test
        2. `appsody operator install --namespace test`
            - this should fail with an error saying "operator exists in namespace test watching namespace test"
        3. `appsody deploy`
            - operator should be installed in namespace default watching namespace default
            - app should be deployed to namespace default
        4. `appsody operator uninstall --namespace test`
            - this should fail with an error saying "unable to uninstall operator with running apps. use --force option to force the uninstall and remove apps"
        5. `appsody operator uninstall`
            - this should fail with an error saying "unable to uninstall operator with running apps. use --force option to force the uninstall and remove apps"
        6. `appsody operator uninstall --namespace test --force`
            - operator should be removed from namespace test
            - app should be removed from namespace test
        7. `appsody operator uninstall --force`
            - operator should be removed from namespace default
            - app should be removed from namespace default

- Test number: 236c
    - Title: New operator install default
    - Issue: 236
    - Prereq:
        - start from new project directory
        - `appsody init <stack>`
    - Steps:
        1. `appsody operator install`
            - operator should be installed in namespace default watching namespace default
        2. `appsody deploy`
            - app should be deployed to namespace default
        3. `appsody deploy delete`
            - app should be deleted from namespace default
        4. `appsody operator uninstall`
            - operator should be deleted from namespace default

- Test number: 236d
    - Title: New operator install custom namespace
    - Issue: 236, 238
    - Prereq:
        - start from new project directory
        - `appsody init <stack>`
        - `kubectl create ns test`
    - Steps:
        1. `appsody operator install --namespace test`
            - operator should be installed in namespace test watching namespace test
        2. `appsody operator uninstall`
            - this should fail with an error saying "no operator exists in namespace default"
        3. `appsody deploy --namespace test`
            - app should be deployed to namespace test
        4. `appsody deploy delete --namespace test2`
            - this should fail with an error saying "namespace test2 does not exist"
        5. `appsody deploy delete --namespace test`
            - app should be deleted from namespace test
        6. `appsody operator uninstall --namespace test`
            - operator should be deleted from namespace test

- Test number: 237a
    - Title: New operator install default namespace custom watchspace
    - Issue: 237, 238, 240
    - Prereq:
        - start from new project directory
        - `appsody init <stack>`
        - `kubectl create ns test`
    - Steps:
        1. `appsody operator install --watchspace test`
            - operator should be installed in namespace default watching namespace test
        2. `appsody operator install --namespace test`
            - this should fail with an error saying "operator exists in namespace default watching namespace test"
        3. `appsody deploy --namespace test`
            - app should deploy to namespace test
        4. `appsody deploy`
            - this should fail with an error saying "existing operator in namespace default is watching namespace test and can not be modified to watch default namespace"
        5. `appsody operator uninstall`
            - this should fail with an error saying "unable to uninstall operator with running apps. use --force option to force the uninstall and remove apps"
        6. `appsody deploy delete --namespace test`
            - app should be removed from namespace test
        7. `appsody operator uninstall`
            - operator should be removed from namespace default
        8. `appsody operator install --watchspace test`
            - operator should be installed in namespace default watching namespace test
        9. `appsody deploy --namespace test`
            - app should deploy to namespace test
       10. `appsody operator uninstall`
            - this should fail with an error saying "unable to uninstall operator with running apps. use --force option to force the uninstall and remove apps"
       11. `appsody operator uninstall --force`
            - operator should be removed from namespace default
            - app should be removed from namespace test

- Test number: 237b
    - Title: New operator install custom namespace custom2 watchspace
    - Issue: 237, 238, 240
    - Prereq:
        - start from new project directory
        - `appsody init <stack>`
        - `kubectl create ns test`
        - `kubectl create ns test2`
    - Steps:
        1. `appsody operator install --namespace test --watchspace test2`
            - operator should be installed in namespace test watching namespace test2
        2. `appsody operator install --namespace test2`
            - this should fail with an error saying "operator exists in namespace test watching namespace test2"
        3. `appsody deploy --namespace test2`
            - app should deploy to namespace test2
            - no new operators should be installed
        4. `appsody deploy --namespace test`
            - this should fail with an error saying "existing operator in namespace test is watching namespace test2 and can not be modified to watch test namespace"
        5. `appsody operator install --namespace test2 --watchspace test`
            - operator should be installed in namespace test2 watching namespace test
        6. `appsody deploy --namespace test`
            - app should be deployed into namespace test
        7. `appsody operator uninstall --namespace test`
            - this should fail with an error saying "unable to uninstall operator with running apps. use --force option to force the uninstall and remove apps"
        8. `appsody operator uninstall --namespace test --force`
            - app in namespace test2 should be removed
            - operator in namespace test should be removed
        9. `appsody operator uninstall`
            - this should fail with an error saying "no operator exists in namespace default"
       10. `appsody operator uninstall --namespace test2`
            - this should fail with an error saying "unable to uninstall operator with running apps. use --force option to force the uninstall and remove apps"
       11. `appsody deploy delete --namespace test`
            - app in namespace test should be removed
       12. `appsody operator uninstall --namespace test2`
            - operator in namespace test2 should be removed

- Test number: 239a
    - Title: New operator install default namespace watchspace all
    - Issue: 239, 240
    - Prereq:
        - start from new project directory
        - `appsody init <stack>`
        - `kubectl create ns test`
        - `kubectl create ns test2`
    - Steps:
        1. `appsody operator install --watchspace --watch-all`
            - operator should be installed in namespace default watching all namespaces
        2. `appsody operator install --namespace test`
           - this should fail with an error saying "operator exists in namespace default watching namespace test"
        3. `appsody deploy`
           - app should be deployed in namespace default
           - no additional operator should be installed
        4. `appsody deploy --namespace test`
           - app should be deployed in namespace test
           - no operator additional operator should be installed
        5. `appsody deploy --namespace test2`
           - app should be deployed in namespace test2
           - no additional operator should be installed
        6. `appsody operator uninstall`
           - this should fail with an error saying "unable to uninstall operator with running apps. use --force option to force the uninstall and remove apps"
        7. `appsody operator uninstall --force`
           - app in namespace default should be removed
           - app in namespace test should be removed
           - app in namespace test2 should be removed
           - operator in namespace default should be removed

- Test number: 239b
    - Title: New operator install custom namespace watchspace all
    - Issue: 239, 240
    - Prereq:
        - start from new project directory
        - `appsody init <stack>`
        - `kubectl create ns test`
        - `kubectl create ns test2`
    - Steps:
        1. `appsody operator install --namespace test --watch-all`
            - operator should be installed in namespace default watching all namespaces
        2. `appsody operator install`
            - this should fail with an error saying "operator exists in namespace test watching namespace default"
        3. `appsody deploy`
            - app should be deployed in namespace default
            - no additional operator should be installed
        4. `appsody deploy --namespace test`
            - app should be deployed in namespace test
            - no additional operator should be installed
        5. `appsody deploy --namespace test2`
            - app should be deployed in namespace test2
            - no additional operator should be installed
        6. `appsody operator uninstall`
            - this should fail with an error saying "no operator exists in namespace default"
        7. `appsody operator uninstall --namespace test`
            - this should fail with an error saying "unable to uninstall operator with running apps. use --force option to force the uninstall and remove apps"
        8. `appsody operator uninstall --force`
            - app in namespace default should be removed
            - app in namespace test should be removed
            - app in namespace test2 should be removed
            - operator in namespace test should be removed

- Test number: 242
    - Title: Limited K8S permissions
    - Issue: 239, 240
    - Prereq:
        - start from new project directory
        - `appsody init <stack>`
        - `kubectl create ns test` (user only has access to namespace test)
    - Steps:
        1. `appsody operator install`
            - this should fail with an error saying "no permission to namespace default"
        2. `appsody operator install --namespace test --watch-all`
            - this should fail with an error saying "no permission to namespace default"
        3. `appsody operator install --namespace test --watchspace "test2"`
            - this should fail with an error saying "no permission to namespace test2"
        4. `appsody deploy`
            - this should fail with an error saying "no permission to namespace default"
        5. `appsody deploy --namespace test`
            - operator should be installed in namespace test watching namespace test
            - app should be deployed in namespace test
        6. `appsody deploy --namespace test2`
            - this should fail with an error saying "no permission to namespace test2"
        7. `appsody deploy delete`
            - this should fail with an error saying "no permission to namespace default"
        8. `appsody operator uninstall`
            - this should fail with an error saying "no permission to namespace default"
        9. `appsody operator uninstall --namespace test2`
            - this should fail with an error saying "no permission to namespace test2"
       10. `appsody deploy delete --namespace test`
            - app should be deleted from namespace test
       11. `appsody deploy --namespace test`
            - apps should be deployed to namespace test
       12. change to different project directory and `appsody init <stack2>`
       13. `appsody deploy --namespace test`
            - app2 should be deployed to namespace test
       14. `appsody operator uninstall --namespace test`
            - this should fail with an error saying "unable to uninstall operator with running apps. use --force option to force the uninstall and remove apps"
       15. `appsody operator uninstall --namespace test --force`
            - app in namespace test should be removed
            - app2 in namespace test should be removed
            - operator in namespace test should be removed

- Test number: 371
    - Title: WATCH_NAMESPACE uses valueFrom
    - Issue: 371
    - Steps:
        1. `kubectl create namespace kabtest`
        2. `kubectl create namespace test`
        3. `kubectl create namespace test1`
        4. `appsody operator install -n kabtest`
            -  Outcome: Operator should install in namespace kabtest
        5. `kubectl edit deploy -n kabtest appsody-operator`
            Change:

                    - name: WATCH_NAMESPACE
                    value: kabtest
            To: 

                    - name: WATCH_NAMESPACE
                    valueFrom:
                        fieldRef:
                        apiVersion: v1
                        fieldPath: metadata.namespace
        6. End the editing session with ESC :wq
        7. `appsody operator install -n test`
            - An operator should be created in namespace test
        8. `appsody operator install -n test1 -w kabtest`
             - This should fail with a message that an operator in namespace kabtest is watching namespace kabtest
        9. `mkdir test`
        10. `cd test`
        11. `appsody init nodejs-express`
        12. `appsody deploy -n kabtest`
            - the application should be created in namespace kabtest
        13. `appsody operator uninstall -n kabtest`
            - The operator in namespace kabtest should fail
        14. `appsody operator uninstall -n kabtest --force`
            - The operator in namespace kabtest should be removed along with the appsody application in kabtest
        15. appsody operator uninstall -n test 
            - The operator in namespace test should be removed

- Test 376:  
    - Title: Test with an operator that is watching multiple namespaces
    - Issue: 376
    - Steps:
        1. `kubectl create namespace kabtest2`
        2. `kubectl create namespace testa`
        3. `kubectl create namespace testa1`
        4. `kubectl create namespace testa2`
        5. appsody operator install -n kabtest2
            - Operator should install in namespace kabtest2
        6. kubectl edit deploy -n kabtest2 appsody-operator
        Change:
        - name: WATCH_NAMESPACE
          value: kabtest2
        To: 
        - name: WATCH_NAMESPACE
          value: kabtest2,testa,testa1
        7. End the editing session with ESC :wq
        8. `appsody operator install -n testa1`
            - This should fail with the message that an operator in namespace kabtest2 is watching namespace,testa
        9. `appsody operator install -n testa1 -w kabtest2`
        - This should fail with a message that an operator in namespace kabtest2 is watching namespace kabtest2,testA,testa1
        10. Create a new directory testa2
        11. cd to testa2
        12. `appsody init nodejs-express`
        13. `appsody deploy -n testa2`
        - The there should be a new operator in namespace testa2 and an appsody application
        14. `appsody deploy -n testa1`
        - should deploy an application in namespace testa1 and no new operator
        15. `appsody deploy delete -n testa2`
        16. `appsody operator uninstall -n testa2`
        - The appsody application and operator should be removed from namespace testa2
        17. `appsody operator uninstall -n kabtest2`
        18. `appsody operator uninstall -n kabtest2 -force`
        - There will be have an kubectl delete failure for rbac.yaml , but that is ok, the reason for this is that we edited the watch spaces outside the appsody operator install code, therefore no cluster rbac gets created.

        [Error] Error from server (NotFound): error when deleting "/Users/kewegnerus.ibm.com/.appsody/deploy/appsody-app-cluster-rbac.yaml": clusterroles.rbac.authorization.k8s.io "appsody-operator-kabtest2" not found

        The operator and crd yamls should be removed 
