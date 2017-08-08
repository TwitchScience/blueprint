angular.module('blueprint.kinesisconfig.create', [
  'ngRoute',
  'blueprint.components.auth',
  'blueprint.components.rest',
  'blueprint.components.store'
]).controller('CreateKinesisConfig', function($scope, $location, $routeParams, Store, KinesisConfig, Auth) {
    $scope.loginName = Auth.getLoginName();
    $scope.isAdmin = Auth.isAdmin();

    $scope.StreamName = '';
    $scope.StreamType = '';
    $scope.AWSAccount = 0;
    $scope.Team = '';
    $scope.Contact = '';
    $scope.Usage = '';
    $scope.Consuminglibrary = '';
    $scope.SpadeConfig = '';
    $scope.configJSON = ''
    $scope.createKinesisConfig = function() {
      Store.clearError();
      try {
        $scope.SpadeConfig = JSON.parse($scope.configJSON)
      } catch (err) {
        Store.setError("Invalid JSON - could not be parsed: " + err)
        return false
      }
      if (!$scope.SpadeConfig.StreamName || !$scope.SpadeConfig.StreamType || $scope.AWSAccount == 0) {
        Store.setError("AWS account, stream name and stream type must be present")
        return false
      }
      KinesisConfig.put({
        "StreamName": $scope.SpadeConfig.StreamName,
        "StreamType": $scope.SpadeConfig.StreamType,
        "AWSAccount": $scope.AWSAccount,
        "Team": $scope.Team,
        "Contact": $scope.Contact,
        "Usage": $scope.Usage,
        "ConsumingLibrary": $scope.ConsumingLibrary,
        "SpadeConfig": $scope.SpadeConfig
      }, function() {
        Store.setMessage("Succesfully created Kinesis config: " + $scope.StreamName)
        $location.path('/kinesisconfigs');
      }, function(err) {
        var msg;
        if (err.data) {
          msg = err.data;
        } else {
          msg = 'Error creating Kinesis Config:' + err;
        }
        Store.setError(msg);
        return false;
      });
    };
  });