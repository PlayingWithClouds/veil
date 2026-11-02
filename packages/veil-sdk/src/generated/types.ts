export default {
    "scalars": [
        1,
        2,
        3,
        5,
        6,
        15,
        16,
        36,
        37,
        39,
        41,
        43,
        44,
        45,
        46,
        52,
        56
    ],
    "types": {
        "Query": {
            "play": [
                7,
                {
                    "url": [
                        1,
                        "String!"
                    ]
                }
            ],
            "videos": [
                19,
                {
                    "filter": [
                        62
                    ],
                    "limit": [
                        2
                    ],
                    "offset": [
                        2
                    ]
                }
            ],
            "videoCount": [
                2,
                {
                    "filter": [
                        62
                    ]
                }
            ],
            "video": [
                19,
                {
                    "id": [
                        3,
                        "ID!"
                    ]
                }
            ],
            "galleries": [
                22,
                {
                    "filter": [
                        62
                    ],
                    "limit": [
                        2
                    ],
                    "offset": [
                        2
                    ]
                }
            ],
            "galleryCount": [
                2,
                {
                    "filter": [
                        62
                    ]
                }
            ],
            "gallery": [
                22,
                {
                    "id": [
                        3,
                        "ID!"
                    ]
                }
            ],
            "performers": [
                24,
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
                24,
                {
                    "id": [
                        3,
                        "ID!"
                    ]
                }
            ],
            "studios": [
                26,
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
                26,
                {
                    "id": [
                        3,
                        "ID!"
                    ]
                }
            ],
            "tags": [
                27,
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
                27,
                {
                    "id": [
                        3,
                        "ID!"
                    ]
                }
            ],
            "plugins": [
                31
            ],
            "plugin": [
                31,
                {
                    "name": [
                        1,
                        "String!"
                    ]
                }
            ],
            "pluginPackages": [
                30,
                {
                    "query": [
                        1
                    ]
                }
            ],
            "trash": [
                35,
                {
                    "kind": [
                        44,
                        "RecordKind!"
                    ],
                    "limit": [
                        2
                    ],
                    "offset": [
                        2
                    ]
                }
            ],
            "jobs": [
                38,
                {
                    "statuses": [
                        37,
                        "[JobStatus!]"
                    ],
                    "kinds": [
                        36,
                        "[JobKind!]"
                    ],
                    "limit": [
                        2
                    ],
                    "offset": [
                        2
                    ]
                }
            ],
            "job": [
                38,
                {
                    "id": [
                        3,
                        "ID!"
                    ]
                }
            ],
            "devices": [
                40
            ],
            "currentDevice": [
                40
            ],
            "storageLocations": [
                42,
                {
                    "deviceId": [
                        3
                    ]
                }
            ],
            "userStates": [
                50,
                {
                    "filter": [
                        51
                    ],
                    "limit": [
                        2
                    ],
                    "offset": [
                        2
                    ]
                }
            ],
            "history": [
                50,
                {
                    "kind": [
                        44
                    ],
                    "limit": [
                        2
                    ],
                    "offset": [
                        2
                    ]
                }
            ],
            "events": [
                53,
                {
                    "filter": [
                        55
                    ],
                    "limit": [
                        2
                    ],
                    "offset": [
                        2
                    ]
                }
            ],
            "eventCount": [
                2,
                {
                    "filter": [
                        55
                    ]
                }
            ],
            "recentSearches": [
                1,
                {
                    "limit": [
                        2
                    ]
                }
            ],
            "savedFilters": [
                63
            ],
            "savedFilter": [
                63,
                {
                    "id": [
                        3,
                        "ID!"
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
                8,
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
                19,
                {
                    "id": [
                        3,
                        "ID!"
                    ]
                }
            ],
            "fetchGallery": [
                22,
                {
                    "id": [
                        3,
                        "ID!"
                    ]
                }
            ],
            "browse": [
                17,
                {
                    "plugin": [
                        1,
                        "String!"
                    ],
                    "kind": [
                        16,
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
                17,
                {
                    "kind": [
                        16,
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
            "editVideo": [
                19,
                {
                    "id": [
                        3,
                        "ID!"
                    ],
                    "edit": [
                        9,
                        "MediaEdit!"
                    ]
                }
            ],
            "editGallery": [
                22,
                {
                    "id": [
                        3,
                        "ID!"
                    ],
                    "edit": [
                        9,
                        "MediaEdit!"
                    ]
                }
            ],
            "editPerformer": [
                24,
                {
                    "id": [
                        3,
                        "ID!"
                    ],
                    "edit": [
                        10,
                        "PerformerEdit!"
                    ]
                }
            ],
            "editStudio": [
                26,
                {
                    "id": [
                        3,
                        "ID!"
                    ],
                    "edit": [
                        11,
                        "StudioEdit!"
                    ]
                }
            ],
            "editTag": [
                27,
                {
                    "id": [
                        3,
                        "ID!"
                    ],
                    "edit": [
                        12,
                        "TagEdit!"
                    ]
                }
            ],
            "lockField": [
                13,
                {
                    "subject": [
                        47,
                        "RecordRefInput!"
                    ],
                    "field": [
                        1,
                        "String!"
                    ],
                    "locked": [
                        5,
                        "Boolean!"
                    ]
                }
            ],
            "revertRevision": [
                5,
                {
                    "id": [
                        3,
                        "ID!"
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
                        34,
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
            "trashRecord": [
                5,
                {
                    "subject": [
                        47,
                        "RecordRefInput!"
                    ]
                }
            ],
            "restoreRecord": [
                5,
                {
                    "subject": [
                        47,
                        "RecordRefInput!"
                    ]
                }
            ],
            "purgeRecord": [
                5,
                {
                    "subject": [
                        47,
                        "RecordRefInput!"
                    ]
                }
            ],
            "emptyTrash": [
                2
            ],
            "queueFetch": [
                38,
                {
                    "subject": [
                        47,
                        "RecordRefInput!"
                    ]
                }
            ],
            "cancelJob": [
                38,
                {
                    "id": [
                        3,
                        "ID!"
                    ]
                }
            ],
            "retryJob": [
                38,
                {
                    "id": [
                        3,
                        "ID!"
                    ]
                }
            ],
            "clearFinishedJobs": [
                2
            ],
            "renameDevice": [
                40,
                {
                    "id": [
                        3,
                        "ID!"
                    ],
                    "name": [
                        1,
                        "String!"
                    ]
                }
            ],
            "addStorageLocation": [
                42,
                {
                    "kind": [
                        41,
                        "StorageKind!"
                    ],
                    "uri": [
                        1,
                        "String!"
                    ],
                    "quotaBytes": [
                        2
                    ]
                }
            ],
            "setStorageQuota": [
                42,
                {
                    "id": [
                        3,
                        "ID!"
                    ],
                    "quotaBytes": [
                        2
                    ]
                }
            ],
            "removeStorageLocation": [
                5,
                {
                    "id": [
                        3,
                        "ID!"
                    ]
                }
            ],
            "rate": [
                50,
                {
                    "subject": [
                        47,
                        "RecordRefInput!"
                    ],
                    "rating": [
                        6
                    ]
                }
            ],
            "setFavorite": [
                50,
                {
                    "subject": [
                        47,
                        "RecordRefInput!"
                    ],
                    "favorite": [
                        5,
                        "Boolean!"
                    ]
                }
            ],
            "setBlocked": [
                50,
                {
                    "subject": [
                        47,
                        "RecordRefInput!"
                    ],
                    "blocked": [
                        5,
                        "Boolean!"
                    ]
                }
            ],
            "setNotes": [
                50,
                {
                    "subject": [
                        47,
                        "RecordRefInput!"
                    ],
                    "notes": [
                        1,
                        "String!"
                    ]
                }
            ],
            "recordEvent": [
                53,
                {
                    "event": [
                        54,
                        "EventInput!"
                    ]
                }
            ],
            "clearEvents": [
                2,
                {
                    "kinds": [
                        52,
                        "[EventKind!]"
                    ]
                }
            ],
            "saveFilter": [
                63,
                {
                    "id": [
                        3
                    ],
                    "name": [
                        1,
                        "String!"
                    ],
                    "filter": [
                        62,
                        "FilterInput!"
                    ],
                    "pinned": [
                        5
                    ]
                }
            ],
            "deleteSavedFilter": [
                5,
                {
                    "id": [
                        3,
                        "ID!"
                    ]
                }
            ],
            "__typename": [
                1
            ]
        },
        "Boolean": {},
        "Float": {},
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
        "MediaEdit": {
            "title": [
                1
            ],
            "details": [
                1
            ],
            "date": [
                1
            ],
            "studioId": [
                3
            ],
            "performerIds": [
                3
            ],
            "tagIds": [
                3
            ],
            "__typename": [
                1
            ]
        },
        "PerformerEdit": {
            "name": [
                1
            ],
            "aliases": [
                1
            ],
            "details": [
                1
            ],
            "links": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "StudioEdit": {
            "name": [
                1
            ],
            "aliases": [
                1
            ],
            "details": [
                1
            ],
            "url": [
                1
            ],
            "parentId": [
                3
            ],
            "__typename": [
                1
            ]
        },
        "TagEdit": {
            "name": [
                1
            ],
            "aliases": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "FieldSource": {
            "field": [
                1
            ],
            "writtenBy": [
                15
            ],
            "source": [
                28
            ],
            "locked": [
                5
            ],
            "updatedAt": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "Revision": {
            "id": [
                3
            ],
            "createdAt": [
                1
            ],
            "subjectKind": [
                44
            ],
            "subjectId": [
                3
            ],
            "field": [
                1
            ],
            "before": [
                43
            ],
            "after": [
                43
            ],
            "writtenBy": [
                15
            ],
            "actorName": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "Actor": {},
        "ListKind": {},
        "Listed": {
            "videos": [
                19
            ],
            "galleries": [
                22
            ],
            "performers": [
                24
            ],
            "studios": [
                26
            ],
            "tags": [
                27
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
                18
            ],
            "status": [
                1
            ],
            "studio": [
                26
            ],
            "performers": [
                24
            ],
            "tags": [
                27
            ],
            "durationSeconds": [
                6
            ],
            "width": [
                2
            ],
            "height": [
                2
            ],
            "vr": [
                20
            ],
            "preview": [
                18
            ],
            "trickplay": [
                21
            ],
            "gallery": [
                22
            ],
            "sources": [
                28
            ],
            "streams": [
                29
            ],
            "userState": [
                50
            ],
            "oCount": [
                2
            ],
            "fieldSources": [
                13
            ],
            "revisions": [
                14,
                {
                    "limit": [
                        2
                    ],
                    "offset": [
                        2
                    ]
                }
            ],
            "__typename": [
                1
            ]
        },
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
                18
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
                6
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
                18
            ],
            "status": [
                1
            ],
            "studio": [
                26
            ],
            "performers": [
                24
            ],
            "tags": [
                27
            ],
            "imageCount": [
                2
            ],
            "images": [
                23
            ],
            "video": [
                19
            ],
            "sources": [
                28
            ],
            "userState": [
                50
            ],
            "oCount": [
                2
            ],
            "fieldSources": [
                13
            ],
            "revisions": [
                14,
                {
                    "limit": [
                        2
                    ],
                    "offset": [
                        2
                    ]
                }
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
                18
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
                18
            ],
            "details": [
                1
            ],
            "attributes": [
                25
            ],
            "links": [
                1
            ],
            "tags": [
                27
            ],
            "sources": [
                28
            ],
            "videoCount": [
                2
            ],
            "galleryCount": [
                2
            ],
            "userState": [
                50
            ],
            "fieldSources": [
                13
            ],
            "revisions": [
                14,
                {
                    "limit": [
                        2
                    ],
                    "offset": [
                        2
                    ]
                }
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
                18
            ],
            "details": [
                1
            ],
            "url": [
                1
            ],
            "parent": [
                26
            ],
            "tags": [
                27
            ],
            "sources": [
                28
            ],
            "videoCount": [
                2
            ],
            "galleryCount": [
                2
            ],
            "userState": [
                50
            ],
            "fieldSources": [
                13
            ],
            "revisions": [
                14,
                {
                    "limit": [
                        2
                    ],
                    "offset": [
                        2
                    ]
                }
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
            "userState": [
                50
            ],
            "revisions": [
                14,
                {
                    "limit": [
                        2
                    ],
                    "offset": [
                        2
                    ]
                }
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
                18
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
                28
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
                32
            ],
            "settingValues": [
                33
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
        },
        "TrashedRecord": {
            "kind": [
                44
            ],
            "id": [
                3
            ],
            "label": [
                1
            ],
            "deletedAt": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "JobKind": {},
        "JobStatus": {},
        "Job": {
            "id": [
                3
            ],
            "kind": [
                36
            ],
            "status": [
                37
            ],
            "subject": [
                48
            ],
            "plugin": [
                1
            ],
            "payload": [
                43
            ],
            "progress": [
                6
            ],
            "error": [
                1
            ],
            "attempts": [
                2
            ],
            "priority": [
                2
            ],
            "device": [
                3
            ],
            "createdAt": [
                1
            ],
            "runAt": [
                1
            ],
            "startedAt": [
                1
            ],
            "finishedAt": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "Platform": {},
        "Device": {
            "id": [
                3
            ],
            "name": [
                1
            ],
            "platform": [
                39
            ],
            "lastSeenAt": [
                1
            ],
            "lastSyncAt": [
                1
            ],
            "current": [
                5
            ],
            "createdAt": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "StorageKind": {},
        "StorageLocation": {
            "id": [
                3
            ],
            "device": [
                40
            ],
            "kind": [
                41
            ],
            "uri": [
                1
            ],
            "quotaBytes": [
                2
            ],
            "usedBytes": [
                2
            ],
            "__typename": [
                1
            ]
        },
        "JSON": {},
        "RecordKind": {},
        "MediaKind": {},
        "MediaStatus": {},
        "RecordRefInput": {
            "kind": [
                44
            ],
            "id": [
                3
            ],
            "__typename": [
                1
            ]
        },
        "RecordRef": {
            "kind": [
                44
            ],
            "id": [
                3
            ],
            "__typename": [
                1
            ]
        },
        "Record": {
            "on_Video": [
                19
            ],
            "on_Gallery": [
                22
            ],
            "on_Performer": [
                24
            ],
            "on_Studio": [
                26
            ],
            "on_Tag": [
                27
            ],
            "__typename": [
                1
            ]
        },
        "UserState": {
            "subjectKind": [
                44
            ],
            "subjectId": [
                3
            ],
            "record": [
                49
            ],
            "rating": [
                6
            ],
            "favorite": [
                5
            ],
            "notes": [
                1
            ],
            "blocked": [
                5
            ],
            "progress": [
                6
            ],
            "viewCount": [
                2
            ],
            "lastViewedAt": [
                1
            ],
            "fields": [
                43
            ],
            "updatedAt": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "UserStateFilter": {
            "kind": [
                44
            ],
            "favorite": [
                5
            ],
            "blocked": [
                5
            ],
            "rated": [
                5
            ],
            "viewed": [
                5
            ],
            "__typename": [
                1
            ]
        },
        "EventKind": {},
        "Event": {
            "id": [
                3
            ],
            "at": [
                1
            ],
            "kind": [
                52
            ],
            "subjectKind": [
                44
            ],
            "subjectId": [
                3
            ],
            "data": [
                43
            ],
            "device": [
                3
            ],
            "sessionId": [
                3
            ],
            "__typename": [
                1
            ]
        },
        "EventInput": {
            "kind": [
                52
            ],
            "subject": [
                47
            ],
            "data": [
                43
            ],
            "sessionId": [
                3
            ],
            "__typename": [
                1
            ]
        },
        "EventFilter": {
            "kinds": [
                52
            ],
            "subject": [
                47
            ],
            "since": [
                1
            ],
            "until": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "SortField": {},
        "Sort": {
            "field": [
                56
            ],
            "descending": [
                5
            ],
            "__typename": [
                1
            ]
        },
        "SortInput": {
            "field": [
                56
            ],
            "descending": [
                5
            ],
            "__typename": [
                1
            ]
        },
        "FieldCondition": {
            "fieldDefId": [
                3
            ],
            "operator": [
                1
            ],
            "value": [
                43
            ],
            "__typename": [
                1
            ]
        },
        "FieldConditionInput": {
            "fieldDefId": [
                3
            ],
            "operator": [
                1
            ],
            "value": [
                43
            ],
            "__typename": [
                1
            ]
        },
        "Filter": {
            "kinds": [
                45
            ],
            "text": [
                1
            ],
            "performerIds": [
                3
            ],
            "studioIds": [
                3
            ],
            "categoryIds": [
                3
            ],
            "includeTagIds": [
                3
            ],
            "excludeTagIds": [
                3
            ],
            "sources": [
                1
            ],
            "dateFrom": [
                1
            ],
            "dateTo": [
                1
            ],
            "minDurationSeconds": [
                2
            ],
            "maxDurationSeconds": [
                2
            ],
            "minRating": [
                6
            ],
            "status": [
                46
            ],
            "inLibrary": [
                5
            ],
            "favorite": [
                5
            ],
            "unwatched": [
                5
            ],
            "vr": [
                5
            ],
            "fields": [
                59
            ],
            "sort": [
                57
            ],
            "__typename": [
                1
            ]
        },
        "FilterInput": {
            "kinds": [
                45
            ],
            "text": [
                1
            ],
            "performerIds": [
                3
            ],
            "studioIds": [
                3
            ],
            "categoryIds": [
                3
            ],
            "includeTagIds": [
                3
            ],
            "excludeTagIds": [
                3
            ],
            "sources": [
                1
            ],
            "dateFrom": [
                1
            ],
            "dateTo": [
                1
            ],
            "minDurationSeconds": [
                2
            ],
            "maxDurationSeconds": [
                2
            ],
            "minRating": [
                6
            ],
            "status": [
                46
            ],
            "inLibrary": [
                5
            ],
            "favorite": [
                5
            ],
            "unwatched": [
                5
            ],
            "vr": [
                5
            ],
            "fields": [
                60
            ],
            "sort": [
                58
            ],
            "__typename": [
                1
            ]
        },
        "SavedFilter": {
            "id": [
                3
            ],
            "name": [
                1
            ],
            "filter": [
                61
            ],
            "pinned": [
                5
            ],
            "createdAt": [
                1
            ],
            "updatedAt": [
                1
            ],
            "__typename": [
                1
            ]
        }
    }
}