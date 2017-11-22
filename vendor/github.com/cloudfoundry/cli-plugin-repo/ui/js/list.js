angular.module('clipr', [])
  .controller('CliprListing',['$scope','$http', function($scope, $http){
      $http({ method: 'GET', url: '/list' }).success(function (data) {
            $scope.list= data;
      }).
      error(function (data) {
          $scope.list={plugins:[]};
      });
  }]);

