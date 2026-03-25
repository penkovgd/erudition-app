package neo4j_test

// prepare sets up a neo4j container and cleans it up after the test.
// func prepare(t *testing.T) *n4jc.Client {
// 	t.Helper()

// 	ctx := context.Background()

// 	_, filename, _, ok := runtime.Caller(0)
// 	if !ok {
// 		t.Fatal("could not get current file path")
// 	}
// 	projectRoot := filepath.Join(filepath.Dir(filename), "../../../../..")

// 	neo4jC, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
// 		ContainerRequest: tc.ContainerRequest{
// 			FromDockerfile: tc.FromDockerfile{
// 				Context:    projectRoot,
// 				Dockerfile: "build/Dockerfile.neo4j",
// 				Tag:        "erudition-app-neo4j",
// 			},
// 			ExposedPorts: []string{"7687/tcp", "7474/tcp"},
// 			Env: map[string]string{
// 				"NEO4J_AUTH":    "neo4j/testpassword",
// 				"NEO4J_PLUGINS": `["apoc"]`,
// 				"NEO4J_server_unmanaged__extension__classes":  "n10s.endpoint=/rdf",
// 				"NEO4J_dbms_security_procedures_unrestricted": "apoc.*,n10s.*",
// 			},
// 			WaitingFor: wait.ForAll(
// 				wait.ForListeningPort("7687/tcp"),
// 				wait.ForLog("Started."),
// 			).WithDeadline(60 * time.Second),
// 		},
// 		Started: true,
// 	})
// 	require.NoError(t, err)

// 	mappedPort, err := neo4jC.MappedPort(ctx, "7687")
// 	require.NoError(t, err)
// 	host, err := neo4jC.Host(ctx)
// 	require.NoError(t, err)

// 	uri := fmt.Sprintf("neo4j://%s:%s", host, mappedPort.Port())

// 	client, err := n4jc.New(ctx, slog.Default(), uri, "neo4j", "testpassword")
// 	require.NoError(t, err)

// 	session := client.NewSession(ctx, neo4j.SessionConfig{})
// 	defer closer.CloseOrLogContext(ctx, slog.Default(), session)

// 	_, err = session.Run(ctx, "CREATE CONSTRAINT n10s_unique_uri IF NOT EXISTS FOR (r:Resource) REQUIRE r.uri IS UNIQUE", nil)
// 	require.NoError(t, err)

// 	_, err = session.Run(ctx, "CALL n10s.graphconfig.init({handleVocabUris: 'MAP', applyNeo4jNaming: true})", nil)
// 	if err != nil {
// 		t.Logf("note: n10s init maybe already done: %v", err)
// 	}

// 	t.Cleanup(func() {
// 		client.Close(ctx)
// 		tc.CleanupContainer(t, neo4jC)
// 	})

// 	return client
// }

// func TestClient_Load(t *testing.T) {
// 	tests := []struct {
// 		name    string
// 		wantErr bool
// 	}{
// 		{name: "with_link", wantErr: false},
// 		{name: "with_literal", wantErr: false},
// 		{name: "with_type", wantErr: false},
// 		{name: "invalid_jsonld", wantErr: true},
// 		{name: "duplicates", wantErr: false},
// 		{name: "mona", wantErr: false},
// 	}

// 	client := prepare(t)

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			ctx := context.Background()

// 			inputData, goldenData := testutils.ReadTestFiles(t, "testdata", tt.name)

// 			// Act
// 			gotErr := client.Load(ctx, inputData)

// 			// Assert
// 			if gotErr != nil {
// 				if !tt.wantErr {
// 					// got error but don't want it
// 					t.Fatal(gotErr)
// 				}
// 				if !strings.Contains(gotErr.Error(), strings.TrimSpace(string(goldenData))) {
// 					// got error and want it - check gotErr and golden
// 					t.Errorf("Load() error mismatch\n got: %s\nwant: %s", gotErr.Error(), goldenData)
// 				}
// 				return
// 			}
// 			if tt.wantErr {
// 				// got no error but want it
// 				t.Fatal("Load() succeeded unexpectedly")
// 			}

// 			result, err := neo4j.ExecuteQuery(ctx, client.Driver,
// 				`MATCH (r:Resource) RETURN r.uri AS uri`,
// 				nil,
// 				neo4j.EagerResultTransformer,
// 				neo4j.ExecuteQueryWithDatabase("neo4j"),
// 			)
// 			require.NoError(t, err)

// 			var got []map[string]string
// 			for _, record := range result.Records {
// 				uri, _ := record.Get("uri")
// 				uriStr, _ := uri.(string)
// 				got = append(got, map[string]string{"uri": uriStr})
// 			}

// 			var want []map[string]string
// 			err = json.Unmarshal(goldenData, &want)
// 			require.NoError(t, err)

// 			if diff := cmp.Diff(want, got); diff != "" {
// 				t.Errorf("Resources mismatch (-want +got):\n%s", diff)
// 			}

// 			t.Cleanup(func() {
// 				_, err := neo4j.ExecuteQuery(ctx, client.Driver,
// 					`MATCH (r:Resource) DETACH DELETE r`,
// 					nil,
// 					neo4j.EagerResultTransformer,
// 					neo4j.ExecuteQueryWithDatabase("neo4j"),
// 				)
// 				assert.NoError(t, err)
// 			})
// 		})
// 	}
// }
