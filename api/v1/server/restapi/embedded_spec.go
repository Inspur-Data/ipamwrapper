// Code generated by go-swagger; DO NOT EDIT.

package restapi

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"encoding/json"
)

var (
	// SwaggerJSON embedded version of the swagger document used at generation time
	SwaggerJSON json.RawMessage
	// FlatSwaggerJSON embedded flattened version of the swagger document used at generation time
	FlatSwaggerJSON json.RawMessage
)

func init() {
	SwaggerJSON = json.RawMessage([]byte(`{
  "consumes": [
    "application/json"
  ],
  "produces": [
    "application/json"
  ],
  "swagger": "2.0",
  "info": {
    "description": "k8-ipam Agent",
    "title": "k8-ipam Agent API",
    "version": "v1"
  },
  "basePath": "/v1",
  "paths": {
    "/healthy": {
      "get": {
        "description": "Check the agent health to make sure whether it's ready\nfor CNI plugin usage\n",
        "tags": [
          "health-check"
        ],
        "summary": "Get health of k8-ipam agent",
        "responses": {
          "200": {
            "description": "Success"
          },
          "500": {
            "description": "Failed"
          }
        }
      }
    },
    "/ipam": {
      "post": {
        "description": "Send a request to k8-ipam to alloc an ip\n",
        "tags": [
          "k8-ipam-agent"
        ],
        "summary": "Alloc ip from k8-ipam",
        "parameters": [
          {
            "name": "ipam-alloc-args",
            "in": "body",
            "required": true,
            "schema": {
              "$ref": "#/definitions/IpamAllocArgs"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "Success",
            "schema": {
              "$ref": "#/definitions/IpamAllocResponse"
            }
          },
          "500": {
            "description": "Allocation failure",
            "schema": {
              "$ref": "#/definitions/Error"
            },
            "x-go-name": "Failure"
          }
        }
      },
      "delete": {
        "description": "Send a request to k8-ipam to delete an ip\n",
        "tags": [
          "k8-ipam-agent"
        ],
        "summary": "Delete ip from k8-ipam",
        "parameters": [
          {
            "name": "ipam-del-args",
            "in": "body",
            "required": true,
            "schema": {
              "$ref": "#/definitions/IpamDelArgs"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "Success"
          },
          "500": {
            "description": "Addresses release failure",
            "schema": {
              "$ref": "#/definitions/Error"
            },
            "x-go-name": "Failure"
          }
        }
      }
    }
  },
  "definitions": {
    "DNS": {
      "description": "k8-ipam DNS",
      "type": "object",
      "properties": {
        "domain": {
          "type": "string"
        },
        "nameservers": {
          "type": "array",
          "items": {
            "type": "string"
          }
        },
        "options": {
          "type": "array",
          "items": {
            "type": "string"
          }
        },
        "search": {
          "type": "array",
          "items": {
            "type": "string"
          }
        }
      }
    },
    "Error": {
      "description": "API error",
      "type": "string"
    },
    "IpConfig": {
      "description": "IP struct",
      "type": "object",
      "required": [
        "version",
        "address",
        "nic"
      ],
      "properties": {
        "address": {
          "type": "string"
        },
        "gateway": {
          "type": "string"
        },
        "ipPool": {
          "type": "string"
        },
        "nic": {
          "type": "string"
        },
        "version": {
          "type": "integer",
          "enum": [
            4,
            6
          ]
        },
        "vlan": {
          "type": "integer"
        }
      }
    },
    "IpamAllocArgs": {
      "description": "Alloc IP request args",
      "type": "object",
      "required": [
        "containerID",
        "ifName",
        "netNamespace",
        "podNamespace",
        "podName"
      ],
      "properties": {
        "containerID": {
          "type": "string"
        },
        "ifName": {
          "type": "string"
        },
        "netNamespace": {
          "type": "string"
        },
        "podName": {
          "type": "string"
        },
        "podNamespace": {
          "type": "string"
        }
      }
    },
    "IpamAllocResponse": {
      "description": "Alloc IP information",
      "type": "object",
      "required": [
        "ip"
      ],
      "properties": {
        "dns": {
          "type": "object",
          "$ref": "#/definitions/DNS"
        },
        "ip": {
          "type": "object",
          "$ref": "#/definitions/IpConfig"
        },
        "route": {
          "type": "object",
          "$ref": "#/definitions/Route"
        }
      }
    },
    "IpamDelArgs": {
      "description": "Delete IP information",
      "type": "object",
      "required": [
        "containerID",
        "ifName",
        "podNamespace",
        "podName"
      ],
      "properties": {
        "containerID": {
          "type": "string"
        },
        "ifName": {
          "type": "string"
        },
        "netNamespace": {
          "type": "string"
        },
        "podName": {
          "type": "string"
        },
        "podNamespace": {
          "type": "string"
        }
      }
    },
    "Route": {
      "description": "k8-ipam Route",
      "type": "object",
      "required": [
        "ifName",
        "dst",
        "gw"
      ],
      "properties": {
        "dst": {
          "type": "string"
        },
        "gw": {
          "type": "string"
        },
        "ifName": {
          "type": "string"
        }
      }
    }
  },
  "x-schemes": [
    "unix"
  ]
}`))
	FlatSwaggerJSON = json.RawMessage([]byte(`{
  "consumes": [
    "application/json"
  ],
  "produces": [
    "application/json"
  ],
  "swagger": "2.0",
  "info": {
    "description": "k8-ipam Agent",
    "title": "k8-ipam Agent API",
    "version": "v1"
  },
  "basePath": "/v1",
  "paths": {
    "/healthy": {
      "get": {
        "description": "Check the agent health to make sure whether it's ready\nfor CNI plugin usage\n",
        "tags": [
          "health-check"
        ],
        "summary": "Get health of k8-ipam agent",
        "responses": {
          "200": {
            "description": "Success"
          },
          "500": {
            "description": "Failed"
          }
        }
      }
    },
    "/ipam": {
      "post": {
        "description": "Send a request to k8-ipam to alloc an ip\n",
        "tags": [
          "k8-ipam-agent"
        ],
        "summary": "Alloc ip from k8-ipam",
        "parameters": [
          {
            "name": "ipam-alloc-args",
            "in": "body",
            "required": true,
            "schema": {
              "$ref": "#/definitions/IpamAllocArgs"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "Success",
            "schema": {
              "$ref": "#/definitions/IpamAllocResponse"
            }
          },
          "500": {
            "description": "Allocation failure",
            "schema": {
              "$ref": "#/definitions/Error"
            },
            "x-go-name": "Failure"
          }
        }
      },
      "delete": {
        "description": "Send a request to k8-ipam to delete an ip\n",
        "tags": [
          "k8-ipam-agent"
        ],
        "summary": "Delete ip from k8-ipam",
        "parameters": [
          {
            "name": "ipam-del-args",
            "in": "body",
            "required": true,
            "schema": {
              "$ref": "#/definitions/IpamDelArgs"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "Success"
          },
          "500": {
            "description": "Addresses release failure",
            "schema": {
              "$ref": "#/definitions/Error"
            },
            "x-go-name": "Failure"
          }
        }
      }
    }
  },
  "definitions": {
    "DNS": {
      "description": "k8-ipam DNS",
      "type": "object",
      "properties": {
        "domain": {
          "type": "string"
        },
        "nameservers": {
          "type": "array",
          "items": {
            "type": "string"
          }
        },
        "options": {
          "type": "array",
          "items": {
            "type": "string"
          }
        },
        "search": {
          "type": "array",
          "items": {
            "type": "string"
          }
        }
      }
    },
    "Error": {
      "description": "API error",
      "type": "string"
    },
    "IpConfig": {
      "description": "IP struct",
      "type": "object",
      "required": [
        "version",
        "address",
        "nic"
      ],
      "properties": {
        "address": {
          "type": "string"
        },
        "gateway": {
          "type": "string"
        },
        "ipPool": {
          "type": "string"
        },
        "nic": {
          "type": "string"
        },
        "version": {
          "type": "integer",
          "enum": [
            4,
            6
          ]
        },
        "vlan": {
          "type": "integer"
        }
      }
    },
    "IpamAllocArgs": {
      "description": "Alloc IP request args",
      "type": "object",
      "required": [
        "containerID",
        "ifName",
        "netNamespace",
        "podNamespace",
        "podName"
      ],
      "properties": {
        "containerID": {
          "type": "string"
        },
        "ifName": {
          "type": "string"
        },
        "netNamespace": {
          "type": "string"
        },
        "podName": {
          "type": "string"
        },
        "podNamespace": {
          "type": "string"
        }
      }
    },
    "IpamAllocResponse": {
      "description": "Alloc IP information",
      "type": "object",
      "required": [
        "ip"
      ],
      "properties": {
        "dns": {
          "type": "object",
          "$ref": "#/definitions/DNS"
        },
        "ip": {
          "type": "object",
          "$ref": "#/definitions/IpConfig"
        },
        "route": {
          "type": "object",
          "$ref": "#/definitions/Route"
        }
      }
    },
    "IpamDelArgs": {
      "description": "Delete IP information",
      "type": "object",
      "required": [
        "containerID",
        "ifName",
        "podNamespace",
        "podName"
      ],
      "properties": {
        "containerID": {
          "type": "string"
        },
        "ifName": {
          "type": "string"
        },
        "netNamespace": {
          "type": "string"
        },
        "podName": {
          "type": "string"
        },
        "podNamespace": {
          "type": "string"
        }
      }
    },
    "Route": {
      "description": "k8-ipam Route",
      "type": "object",
      "required": [
        "ifName",
        "dst",
        "gw"
      ],
      "properties": {
        "dst": {
          "type": "string"
        },
        "gw": {
          "type": "string"
        },
        "ifName": {
          "type": "string"
        }
      }
    }
  },
  "x-schemes": [
    "unix"
  ]
}`))
}