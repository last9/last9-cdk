"use strict";
var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    function adopt(value) { return value instanceof P ? value : new P(function (resolve) { resolve(value); }); }
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
var __generator = (this && this.__generator) || function (thisArg, body) {
    var _ = { label: 0, sent: function() { if (t[0] & 1) throw t[1]; return t[1]; }, trys: [], ops: [] }, f, y, t, g;
    return g = { next: verb(0), "throw": verb(1), "return": verb(2) }, typeof Symbol === "function" && (g[Symbol.iterator] = function() { return this; }), g;
    function verb(n) { return function (v) { return step([n, v]); }; }
    function step(op) {
        if (f) throw new TypeError("Generator is already executing.");
        while (g && (g = 0, op[0] && (_ = 0)), _) try {
            if (f = 1, y && (t = op[0] & 2 ? y["return"] : op[0] ? y["throw"] || ((t = y["return"]) && t.call(y), 0) : y.next) && !(t = t.call(y, op[1])).done) return t;
            if (y = 0, t) op = [op[0] & 2, t.value];
            switch (op[0]) {
                case 0: case 1: t = op; break;
                case 4: _.label++; return { value: op[1], done: false };
                case 5: _.label++; y = op[1]; op = [0]; continue;
                case 7: op = _.ops.pop(); _.trys.pop(); continue;
                default:
                    if (!(t = _.trys, t = t.length > 0 && t[t.length - 1]) && (op[0] === 6 || op[0] === 2)) { _ = 0; continue; }
                    if (op[0] === 3 && (!t || (op[1] > t[0] && op[1] < t[3]))) { _.label = op[1]; break; }
                    if (op[0] === 6 && _.label < t[1]) { _.label = t[1]; t = op; break; }
                    if (t && _.label < t[2]) { _.label = t[2]; _.ops.push(op); break; }
                    if (t[2]) _.ops.pop();
                    _.trys.pop(); continue;
            }
            op = body.call(thisArg, _);
        } catch (e) { op = [6, e]; y = 0; } finally { f = t = 0; }
        if (op[0] & 5) throw op[1]; return { value: op[0] ? op[1] : void 0, done: true };
    }
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.getPromClient = void 0;
var fs = require("fs");
var path = require("path");
var os = require("os");
var promClient = require("prom-client");
var getPackageJson = function () {
    var packageJsonPath = path.join(process.cwd(), 'package.json');
    try {
        var packageJson = fs.readFileSync(packageJsonPath, 'utf-8');
        return JSON.parse(packageJson);
    }
    catch (error) {
        console.error('Error parsing package.json');
        return null;
    }
};
var getHostIpAddress = function () {
    var networkInterfaces = os.networkInterfaces();
    // Iterate over network interfaces to find a non-internal IPv4 address
    for (var interfaceName in networkInterfaces) {
        var interfaces = networkInterfaces[interfaceName];
        if (interfaces) {
            for (var _i = 0, interfaces_1 = interfaces; _i < interfaces_1.length; _i++) {
                var iface = interfaces_1[_i];
                // Skip internal and non-IPv4 addresses
                if (!iface.internal && iface.family === 'IPv4') {
                    return iface.address;
                }
            }
        }
    }
    // Return null if no IP address is found
    return null;
};
var getPromClient = function (_a) {
    var environment = _a.environment;
    return __awaiter(void 0, void 0, void 0, function () {
        var _b, version, name, ip;
        return __generator(this, function (_c) {
            _b = getPackageJson(), version = _b.version, name = _b.name;
            ip = getHostIpAddress();
            promClient.register.setDefaultLabels({
                program: name,
                version: version,
                environment: environment,
                host: os.hostname(),
                ip: ip,
            });
            return [2 /*return*/, promClient];
        });
    });
};
exports.getPromClient = getPromClient;
