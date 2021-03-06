angular.module('blueprint.schema.create', [
  'ngRoute',
  'blueprint.components.auth',
  'blueprint.components.column',
  'blueprint.components.rest',
  'blueprint.components.store'
]).controller('CreateSchema', function($scope, $location, $q, $routeParams, Store, Schema, Types, Suggestions, Column, Auth, EventMetadata) {
    $scope.loginName = Auth.getLoginName();
    Auth.globalIsEditable($scope);
    $scope.datastores = {
      "ace": true,
      "tahoe": false
    }
    var types, suggestions, suggestionData;
    var typeData = Types.get(function(data) {
      if (data) {
        types = data.result;
      } else {
        Store.setError('Failed to fetch type information', undefined)
        types = [];
      }
    }).$promise;

    if ($routeParams['scope']) {
      suggestionData = Suggestions.get($routeParams, function(data) {
        if (data) {
          suggestions = data;
        }
      }).$promise;
    } else {
      var deferScratch = $q.defer();
      deferScratch.resolve();
      suggestionData = deferScratch.promise;
    }

    var rewriteColumns = function(cols) {
      var rewrites = [
        {"Name": "app_version", "Change": [["size", 32]]},
        {"Name": "browser", "Change": [["size", 255]]},
        {"Name": "channel", "Change": [["size", 25]]},
        {"Name": "channel_id", "Change": [["Transformer", "userIDWithMapping"], ["mappingColumn", "channel"]]},
        {"Name": "content_mode", "Change": [["size", 32]]},
        {"Name": "device_id", "Change": [["size", 32]]},
        {"Name": "domain", "Change": [["size", 255]]},
        {"Name": "game", "Change": [["size", 64]]},
        {"Name": "host_channel", "Change": [["size", 25]]},
        {"Name": "host_channel_id", "Change": [["Transformer", "userIDWithMapping"], ["mappingColumn", "host_channel"]]},
        {"Name": "language", "Change": [["size", 8]]},
        {"Name": "login", "Change": [["size", 25]]},
        {"Name": "platform", "Change": [["size", 40]]},
        {"Name": "player", "Change": [["size", 32]]},
        {"Name": "preferred_language", "Change": [["size", 8]]},
        {"Name": "received_language", "Change": [["size", 8]]},
        {"Name": "referrer_domain", "Change": [["size", 255]]},
        {"Name": "referrer_url", "Change": [["size", 255]]},
        {"Name": "url", "Change": [["size", 255]]},
        {"Name": "user_agent", "Change": [["size", 255]]},
        {"Name": "user_id", "Change": [["Transformer", "userIDWithMapping"], ["mappingColumn", "login"]]},
        {"Name": "vod_id", "Change": [["size", 16]]},
      ];

      var deletes = [
        "token",
      ];

      angular.forEach(rewrites, function (rule) {
        angular.forEach(cols, function(col) {
          if (col.InboundName == rule.Name) {
            angular.forEach(rule.Change, function(change) {
              col[change[0]] = change[1];
            })
          }
        });
      });

      angular.forEach(deletes, function (d) {
        for (i=0; i<cols.length; i++) {
          if (cols[i].InboundName == d) {
            cols.splice(i, 1);
            break;
          }
        }
      });
    };

    $q.all([typeData, suggestionData]).then(function() {
      var event = {distkey:''};
      var defaultColumns = [{
          InboundName: 'time',
          OutboundName: 'time',
          Transformer: 'f@timestamp@unix',
          ColumnCreationOptions: ' sortkey',
          mappingColumn: ''
        },{
          InboundName: 'time',
          OutboundName: 'time_utc',
          Transformer: 'f@timestamp@unix-utc',
          ColumnCreationOptions: '',
          mappingColumn: ''
        },{
          InboundName: 'ip',
          OutboundName: 'ip',
          Transformer: 'varchar',
          size: 15,
          ColumnCreationOptions: '',
          mappingColumn: ''
        },{
          InboundName: 'ip',
          OutboundName: 'city',
          Transformer: 'ipCity',
          ColumnCreationOptions: '',
          mappingColumn: ''
        },{
          InboundName: 'ip',
          OutboundName: 'country',
          Transformer: 'ipCountry',
          ColumnCreationOptions: '',
          mappingColumn: ''
        },{
          InboundName: 'ip',
          OutboundName: 'region',
          Transformer: 'ipRegion',
          ColumnCreationOptions: '',
          mappingColumn: ''
        },{
          InboundName: 'ip',
          OutboundName: 'asn_id',
          Transformer: 'ipAsnInteger',
          ColumnCreationOptions: '',
          mappingColumn: ''
        }];
      // this is icky, it is tightly coupled to what spade is
      // looking for. It would be good to have an intermediate
      // representation which BluePrint converts to what spade cares
      // about but for the timebeing this is the quickest solution
      if (!suggestions) {
        event.Columns = defaultColumns;
      } else {
        event = suggestions;
        event.Columns.sort(function(a, b) {return b.OccurrenceProbability - a.OccurrenceProbability});

        for (i=0; i<event.Columns.length; i++) {
          if (event.Columns[i].InboundName == 'time') {
            event.Columns.splice(i, 1);
            break;
          }
        }

        var re = /\((\d+)\)/
        angular.forEach(event.Columns, function(col) {
          if (col.Transformer == 'varchar') {
            var match = re.exec(col.ColumnCreationOptions);
            if (match) {
              col.size = parseInt(match[1]);
            }
            col.ColumnCreationOptions = '';
          }
          if (col.InboundName == 'device_id') {
            event.distkey = 'device_id';
          }
        });

        event.Columns = defaultColumns.concat(event.Columns);
        rewriteColumns(event.Columns);
      }

      $scope.event = event;
      $scope.event.EventName = "";
      $scope.types = types;
      $scope.newCol = Column.make();
      $scope.usingMappingTransformer = Column.usingMappingTransformer;
      $scope.validInboundNames = function() {
        var inboundNames = {};
        angular.forEach($scope.event.Columns, function(col){
          inboundNames[col.InboundName] = true;
        });
        return Object.keys(inboundNames);
      };
      $scope.addColumnToSchema = function(column) {
        if (!Column.validate(column)) {
          Store.setError("New column is invalid", undefined);
          return false;
        }
        Store.clearError();
        $scope.event.Columns.push(column);
        $scope.newCol = Column.make();
        document.getElementById('newInboundName').focus();
      };
      $scope.dropColumnFromSchema = function(columnInd) {
        $scope.event.Columns.splice(columnInd, 1);
      }
      $scope.outboundNameBlacklist = ["date"];
      $scope.createSchema = function() {
        Store.clearError();
        var setDistKey = $scope.event.distkey;
        var nameSet = {};
        var inboundNames = $scope.validInboundNames();
        var hasValidTime = false;
        var eventNameLength = $scope.event.EventName.length;

        if (eventNameLength < 1 || eventNameLength > 127) {
          Store.setError("Event name must be between 1 and 127 characters, given length " + eventNameLength);
        }
        angular.forEach($scope.event.Columns, function(item) {
          if(item.OutboundName == "time" && item.InboundName == "time" && item.Transformer == "f@timestamp@unix"){
            hasValidTime = true;
          }

          if($scope.outboundNameBlacklist.indexOf(item.OutboundName.toLowerCase()) != -1){
            Store.setError("Cannot have outbound name '"+item.OutboundName+"'. It is a reserved identifier.");
          }

          if(item.OutboundName in nameSet){
            Store.setError("Cannot repeat column name. Repeated '"+item.OutboundName+"'");
            return false;
          } else {
            nameSet[item.OutboundName] = true;
          }
          if (!Column.validate(item)) {
            Store.setError("At least one column is invalid; look at '" + item.InboundName + "'", undefined);
            return false;
          }

          if (Column.usingMappingTransformer(item)) {
            if (!item.mappingColumn) {
              Store.setError("Column '" + item.OutboundName + "' is invalid (needs nonempty mapping column)");
              return false;
            }
            if (item.mappingColumn === item.InboundName) {
              Store.setError("Cannot use a column for its own mapping. Column with problem: " + item.OutboundName);
              return false;
            }
            if (inboundNames.indexOf(item.mappingColumn) == -1) {
              Store.setError("Can't add a column using a mapping that is not in the schema. Offending name: " + item.OutboundName);
              return false;
            }
            item.SupportingColumns = item.mappingColumn;
          }

          if (!item.ColumnCreationOptions) {
            item.ColumnCreationOptions = '';
          }
          if (item.Transformer === 'varchar') {
            item.ColumnCreationOptions = '(' + item.size + ')';
          }
          if (setDistKey == item.OutboundName) {
            item.ColumnCreationOptions += ' distkey';
          }
          if (item.Transformer === 'int') {
            item.Transformer = 'bigint';
          }
        });
        if(!hasValidTime) {
          Store.setError("Must have time->time of type f@timestamp@unix.");
        }

        if (Store.getError()) {
          return;
        }
        delete $scope.event.distkey;

        var datastoreValue = "";
        var datastores = Object.keys($scope.datastores).filter(function(datastore) {
          return $scope.datastores[datastore];
        });
        if (datastores.length) {
          datastoreValue = datastores.join(",");
        }

        Schema.put($scope.event, function() {
          EventMetadata.update(
            {event: $scope.event.EventName},
            {MetadataType: "datastores",
             MetadataValue: datastoreValue,
            },
            function() {
              Store.setMessage("Successfully created schema: " + $scope.event.EventName);
              $location.path('/schema/' + $scope.event.EventName);
            },
            function(err) {
              var msg = "Error saving target datastores, please try to update it below";
              Store.setError(msg);
              $location.path('/schema/' + $scope.event.EventName);
            });
        }, function(err) {
          var msg;
          if (err.data) {
            msg = err.data;
          } else {
            msg = 'Error creating schema:' + err;
          }
          Store.setError(msg, '/schemas');
        });
      };
    });
  });
