export default {
    "scalars": [
        1,
        2,
        3,
        5,
        9,
        13
    ],
    "types": {
        "Query": {
            "play": [
                6,
                {
                    "url": [
                        1,
                        "String!"
                    ]
                }
            ],
            "videos": [
                12,
                {
                    "filter": [
                        8
                    ],
                    "limit": [
                        2
                    ],
                    "offset": [
                        2
                    ]
                }
            ],
            "video": [
                12,
                {
                    "id": [
                        3,
                        "ID!"
                    ]
                }
            ],
            "galleries": [
                16,
                {
                    "filter": [
                        8
                    ],
                    "limit": [
                        2
                    ],
                    "offset": [
                        2
                    ]
                }
            ],
            "gallery": [
                16,
                {
                    "id": [
                        3,
                        "ID!"
                    ]
                }
            ],
            "performers": [
                18,
                {
                    "text": [
                        1
                    ],
                    "limit": [
                        2
                    ],
                    "offset": [
                        2
                    ]
                }
            ],
            "performer": [
                18,
                {
                    "id": [
                        3,
                        "ID!"
                    ]
                }
            ],
            "studios": [
                20,
                {
                    "text": [
                        1
                    ],
                    "limit": [
                        2
                    ],
                    "offset": [
                        2
                    ]
                }
            ],
            "studio": [
                20,
                {
                    "id": [
                        3,
                        "ID!"
                    ]
                }
            ],
            "tags": [
                21,
                {
                    "text": [
                        1
                    ],
                    "limit": [
                        2
                    ],
                    "offset": [
                        2
                    ]
                }
            ],
            "tag": [
                21,
                {
                    "id": [
                        3,
                        "ID!"
                    ]
                }
            ],
            "plugins": [
                25
            ],
            "plugin": [
                25,
                {
                    "name": [
                        1,
                        "String!"
                    ]
                }
            ],
            "pluginPackages": [
                24,
                {
                    "query": [
                        1
                    ]
                }
            ],
            "__typename": [
                1
            ]
        },
        "String": {},
        "Int": {},
        "ID": {},
        "Mutation": {
            "scrape": [
                7,
                {
                    "plugin": [
                        1,
                        "String!"
                    ],
                    "url": [
                        1,
                        "String!"
                    ]
                }
            ],
            "fetchVideo": [
                12,
                {
                    "id": [
                        3,
                        "ID!"
                    ]
                }
            ],
            "fetchGallery": [
                16,
                {
                    "id": [
                        3,
                        "ID!"
                    ]
                }
            ],
            "browse": [
                10,
                {
                    "plugin": [
                        1,
                        "String!"
                    ],
                    "kind": [
                        9,
                        "ListKind!"
                    ],
                    "query": [
                        1
                    ],
                    "limit": [
                        2
                    ],
                    "offset": [
                        2
                    ]
                }
            ],
            "search": [
                10,
                {
                    "kind": [
                        9,
                        "ListKind!"
                    ],
                    "query": [
                        1,
                        "String!"
                    ],
                    "limit": [
                        2
                    ]
                }
            ],
            "setPluginEnabled": [
                5,
                {
                    "name": [
                        1,
                        "String!"
                    ],
                    "enabled": [
                        5,
                        "Boolean!"
                    ]
                }
            ],
            "updatePluginSettings": [
                5,
                {
                    "plugin": [
                        1,
                        "String!"
                    ],
                    "values": [
                        28,
                        "[PluginSettingValueInput!]!"
                    ]
                }
            ],
            "installPlugin": [
                5,
                {
                    "packageName": [
                        1,
                        "String!"
                    ]
                }
            ],
            "uninstallPlugin": [
                5,
                {
                    "name": [
                        1,
                        "String!"
                    ]
                }
            ],
            "updatePlugins": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "Boolean": {},
        "Playable": {
            "url": [
                1
            ],
            "mimeType": [
                1
            ],
            "quality": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "Scraped": {
            "kind": [
                1
            ],
            "id": [
                3
            ],
            "__typename": [
                1
            ]
        },
        "MediaFilter": {
            "text": [
                1
            ],
            "performerId": [
                3
            ],
            "studioId": [
                3
            ],
            "tagId": [
                3
            ],
            "__typename": [
                1
            ]
        },
        "ListKind": {},
        "Listed": {
            "videos": [
                12
            ],
            "galleries": [
                16
            ],
            "performers": [
                18
            ],
            "studios": [
                20
            ],
            "tags": [
                21
            ],
            "__typename": [
                1
            ]
        },
        "Asset": {
            "uri": [
                1
            ],
            "mimeType": [
                1
            ],
            "width": [
                2
            ],
            "height": [
                2
            ],
            "blurhash": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "Video": {
            "id": [
                3
            ],
            "createdAt": [
                1
            ],
            "updatedAt": [
                1
            ],
            "title": [
                1
            ],
            "details": [
                1
            ],
            "date": [
                1
            ],
            "cover": [
                11
            ],
            "status": [
                1
            ],
            "studio": [
                20
            ],
            "performers": [
                18
            ],
            "tags": [
                21
            ],
            "durationSeconds": [
                13
            ],
            "width": [
                2
            ],
            "height": [
                2
            ],
            "vr": [
                14
            ],
            "preview": [
                11
            ],
            "trickplay": [
                15
            ],
            "gallery": [
                16
            ],
            "sources": [
                22
            ],
            "streams": [
                23
            ],
            "__typename": [
                1
            ]
        },
        "Float": {},
        "VrFormat": {
            "projection": [
                1
            ],
            "stereo": [
                1
            ],
            "fieldOfView": [
                2
            ],
            "__typename": [
                1
            ]
        },
        "Trickplay": {
            "sheets": [
                11
            ],
            "columns": [
                2
            ],
            "rows": [
                2
            ],
            "tileWidth": [
                2
            ],
            "tileHeight": [
                2
            ],
            "intervalSeconds": [
                13
            ],
            "__typename": [
                1
            ]
        },
        "Gallery": {
            "id": [
                3
            ],
            "createdAt": [
                1
            ],
            "updatedAt": [
                1
            ],
            "title": [
                1
            ],
            "details": [
                1
            ],
            "date": [
                1
            ],
            "cover": [
                11
            ],
            "status": [
                1
            ],
            "studio": [
                20
            ],
            "performers": [
                18
            ],
            "tags": [
                21
            ],
            "imageCount": [
                2
            ],
            "images": [
                17
            ],
            "video": [
                12
            ],
            "sources": [
                22
            ],
            "__typename": [
                1
            ]
        },
        "Image": {
            "id": [
                3
            ],
            "position": [
                2
            ],
            "asset": [
                11
            ],
            "width": [
                2
            ],
            "height": [
                2
            ],
            "__typename": [
                1
            ]
        },
        "Performer": {
            "id": [
                3
            ],
            "createdAt": [
                1
            ],
            "updatedAt": [
                1
            ],
            "name": [
                1
            ],
            "aliases": [
                1
            ],
            "image": [
                11
            ],
            "details": [
                1
            ],
            "attributes": [
                19
            ],
            "links": [
                1
            ],
            "tags": [
                21
            ],
            "sources": [
                22
            ],
            "videoCount": [
                2
            ],
            "galleryCount": [
                2
            ],
            "__typename": [
                1
            ]
        },
        "PerformerAttributes": {
            "gender": [
                1
            ],
            "birthdate": [
                1
            ],
            "deathDate": [
                1
            ],
            "country": [
                1
            ],
            "ethnicity": [
                1
            ],
            "eyeColor": [
                1
            ],
            "hairColor": [
                1
            ],
            "heightCm": [
                2
            ],
            "weightKg": [
                2
            ],
            "measurements": [
                1
            ],
            "fakeTits": [
                1
            ],
            "tattoos": [
                1
            ],
            "piercings": [
                1
            ],
            "careerStart": [
                2
            ],
            "careerEnd": [
                2
            ],
            "__typename": [
                1
            ]
        },
        "Studio": {
            "id": [
                3
            ],
            "createdAt": [
                1
            ],
            "updatedAt": [
                1
            ],
            "name": [
                1
            ],
            "aliases": [
                1
            ],
            "image": [
                11
            ],
            "details": [
                1
            ],
            "url": [
                1
            ],
            "parent": [
                20
            ],
            "tags": [
                21
            ],
            "sources": [
                22
            ],
            "videoCount": [
                2
            ],
            "galleryCount": [
                2
            ],
            "__typename": [
                1
            ]
        },
        "Tag": {
            "id": [
                3
            ],
            "name": [
                1
            ],
            "aliases": [
                1
            ],
            "videoCount": [
                2
            ],
            "galleryCount": [
                2
            ],
            "__typename": [
                1
            ]
        },
        "Source": {
            "id": [
                3
            ],
            "plugin": [
                1
            ],
            "externalId": [
                1
            ],
            "url": [
                1
            ],
            "title": [
                1
            ],
            "image": [
                11
            ],
            "fetchedAt": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "Stream": {
            "id": [
                3
            ],
            "url": [
                1
            ],
            "mimeType": [
                1
            ],
            "quality": [
                1
            ],
            "height": [
                2
            ],
            "speedKbps": [
                2
            ],
            "source": [
                22
            ],
            "__typename": [
                1
            ]
        },
        "PluginPackage": {
            "name": [
                1
            ],
            "version": [
                1
            ],
            "description": [
                1
            ],
            "installedVersion": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "Plugin": {
            "name": [
                1
            ],
            "packageName": [
                1
            ],
            "localBuild": [
                5
            ],
            "displayName": [
                1
            ],
            "iconUrl": [
                1
            ],
            "description": [
                1
            ],
            "version": [
                1
            ],
            "capabilities": [
                1
            ],
            "domains": [
                1
            ],
            "content": [
                1
            ],
            "enabled": [
                5
            ],
            "requiresSolver": [
                5
            ],
            "available": [
                5
            ],
            "settings": [
                26
            ],
            "settingValues": [
                27
            ],
            "__typename": [
                1
            ]
        },
        "PluginSettingField": {
            "key": [
                1
            ],
            "label": [
                1
            ],
            "description": [
                1
            ],
            "type": [
                1
            ],
            "required": [
                5
            ],
            "default": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "PluginSettingValue": {
            "key": [
                1
            ],
            "value": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "PluginSettingValueInput": {
            "key": [
                1
            ],
            "value": [
                1
            ],
            "__typename": [
                1
            ]
        }
    }
}