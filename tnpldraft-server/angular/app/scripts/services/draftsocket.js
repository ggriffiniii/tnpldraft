'use strict';

angular.module('tnpldraftApp')
  .factory('draftSocket', ['$rootScope', '$interval', function ($scope, $interval) {
    var socket = new WebSocket('ws://psh.randomhost.net:8082/ws/5');

		var service = {};

		socket.onmessage = function(e) {
			$scope.$apply(function() {
				var msg = angular.fromJson(e.data);
				if (msg.type === 'TimeResponse') {
					++service.timeSamples;
					var now = +new Date();
					var estimatedServerSendTime = (now - service.lastSyncSendTime) / 2 + now;
					var serverTime = +new Date(msg.data.time);
					var serverOffset = serverTime - estimatedServerSendTime;
					var movingSampleCount = Math.min(5, service.timeSamples);
					service.serverOffset = ((movingSampleCount - 1) * service.serverOffset + serverOffset) / movingSampleCount;
				} else if (service.onmessage) {
					service.onmessage(e);
				}
			});
		}

		function syncTimeWithServer() {
			service.lastSyncSendTime = +new Date();
			socket.send(angular.toJson({type: 'TimeRequest'}));
		}

		socket.onopen = function() {
			service.serverOffset = 0;
			service.timeSamples = 0;
			syncTimeWithServer();
			$interval(syncTimeWithServer, 5000);
		}

		service.send = function() {
			return socket.send.apply(socket, arguments);
		}

		service.serverTime = function() {
			var now = +new Date();
			return new Date(now + service.serverOffset);
		}

    // Public API here
    return service;
  }]);
