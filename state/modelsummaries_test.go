// Copyright 2017 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package state_test

import (
	"sort"
	"time"

	gc "gopkg.in/check.v1"
	"gopkg.in/juju/names.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/juju/juju/instance"
	"github.com/juju/juju/permission"
	"github.com/juju/juju/state"
	"github.com/juju/juju/status"
	"github.com/juju/juju/testing/factory"
	jc "github.com/juju/testing/checkers"
)

type ModelSummariesSuite struct {
	ConnSuite
}

var _ = gc.Suite(&ModelSummariesSuite{})

func (s *ModelSummariesSuite) Setup4Models(c *gc.C) map[string]string {
	modelUUIDs := make(map[string]string)
	user1 := s.Factory.MakeUser(c, &factory.UserParams{
		Name:        "user1write",
		NoModelUser: true,
	})
	st1 := s.Factory.MakeModel(c, &factory.ModelParams{
		Name:  "user1model",
		Owner: user1.Tag(),
	})
	modelUUIDs["user1model"] = st1.ModelUUID()
	st1.Close()
	user2 := s.Factory.MakeUser(c, &factory.UserParams{
		Name:        "user2read",
		NoModelUser: true,
	})
	st2 := s.Factory.MakeModel(c, &factory.ModelParams{
		Name:  "user2model",
		Owner: user2.Tag(),
	})
	modelUUIDs["user2model"] = st2.ModelUUID()
	st2.Close()
	user3 := s.Factory.MakeUser(c, &factory.UserParams{
		Name:        "user3admin",
		NoModelUser: true,
	})
	st3 := s.Factory.MakeModel(c, &factory.ModelParams{
		Name:  "user3model",
		Owner: user3.Tag(),
	})
	modelUUIDs["user3model"] = st3.ModelUUID()
	st3.Close()
	owner := s.Model.Owner()
	sharedSt := s.Factory.MakeModel(c, &factory.ModelParams{
		Name: "shared",
		// Owned by test-admin
		Owner: owner,
	})
	modelUUIDs["shared"] = sharedSt.ModelUUID()
	defer sharedSt.Close()
	sharedModel, err := sharedSt.Model()
	c.Assert(err, jc.ErrorIsNil)
	_, err = sharedModel.AddUser(state.UserAccessSpec{
		User:      user1.UserTag(),
		CreatedBy: owner,
		Access:    "write",
	})
	c.Assert(err, jc.ErrorIsNil)
	// User 2 has read access to the shared model
	_, err = sharedModel.AddUser(state.UserAccessSpec{
		User:      user2.UserTag(),
		CreatedBy: owner,
		Access:    "read",
	})
	c.Assert(err, jc.ErrorIsNil)
	_, err = sharedModel.AddUser(state.UserAccessSpec{
		User:      user3.UserTag(),
		CreatedBy: owner,
		Access:    "admin",
	})
	c.Assert(err, jc.ErrorIsNil)
	return modelUUIDs
}

func (s *ModelSummariesSuite) modelNamesForUser(c *gc.C, user string) []string {
	tag := names.NewUserTag(user)
	isSuper, err := s.State.IsUserSuperuser(tag)
	c.Assert(err, jc.ErrorIsNil)
	modelQuery, closer, err := s.State.ModelQueryForUser(tag, isSuper)
	defer closer()
	c.Assert(err, jc.ErrorIsNil)
	var docs []struct {
		Name string `bson:"name"`
	}
	modelQuery.Select(bson.M{"name": 1})
	err = modelQuery.All(&docs)
	c.Assert(err, jc.ErrorIsNil)
	names := make([]string, 0)
	for _, doc := range docs {
		names = append(names, doc.Name)
	}
	sort.Strings(names)
	return names
}

func (s *ModelSummariesSuite) TestModelsForUserAdmin(c *gc.C) {
	s.Setup4Models(c)
	names := s.modelNamesForUser(c, s.Model.Owner().Name())
	// Admin always gets to see all models
	c.Check(names, gc.DeepEquals, []string{"shared", "testenv", "user1model", "user2model", "user3model"})
}

func (s *ModelSummariesSuite) TestModelsForUser1(c *gc.C) {
	// User1 is only added to the model they own and the shared model
	s.Setup4Models(c)
	names := s.modelNamesForUser(c, "user1write")
	// Admin always gets to see all models
	c.Check(names, gc.DeepEquals, []string{"shared", "user1model"})
}

func (s *ModelSummariesSuite) TestModelsForUser2(c *gc.C) {
	// User1 is only added to the model they own and the shared model
	s.Setup4Models(c)
	names := s.modelNamesForUser(c, "user2read")
	// Admin always gets to see all models
	c.Check(names, gc.DeepEquals, []string{"shared", "user2model"})
}

func (s *ModelSummariesSuite) TestModelsForUser3(c *gc.C) {
	s.Setup4Models(c)
	names := s.modelNamesForUser(c, "user3admin")
	// Admin always gets to see all models
	c.Check(names, gc.DeepEquals, []string{"shared", "user3model"})
}

func (s *ModelSummariesSuite) TestContainsConfigInformation(c *gc.C) {
	s.Setup4Models(c)
	summaries, err := s.State.ModelSummariesForUser(names.NewUserTag("user1write"))
	c.Assert(err, jc.ErrorIsNil)
	c.Check(summaries, gc.HasLen, 2)
	// We don't guarantee the order of the summaries, but the data for each model should match the same
	// information you would get if you instantiate the model directly
	summaryA := summaries[0]
	model, closer, err := s.StatePool.GetModel(summaryA.UUID)
	defer closer()
	c.Assert(err, jc.ErrorIsNil)
	conf, err := model.Config()
	c.Assert(err, jc.ErrorIsNil)
	c.Check(summaryA.ProviderType, gc.Equals, conf.Type())
	version, ok := conf.AgentVersion()
	c.Assert(ok, jc.IsTrue)
	c.Check(summaryA.AgentVersion, gc.NotNil)
	c.Check(*summaryA.AgentVersion, gc.Equals, version)
	series, ok := conf.DefaultSeries()
	c.Assert(ok, jc.IsTrue)
	c.Check(summaryA.DefaultSeries, gc.Equals, series)
}

func (s *ModelSummariesSuite) TestContainsProviderType(c *gc.C) {
	s.Setup4Models(c)
	summaries, err := s.State.ModelSummariesForUser(names.NewUserTag("user1write"))
	c.Assert(err, jc.ErrorIsNil)
	c.Check(summaries, gc.HasLen, 2)
	// We don't guarantee the order of the summaries, but both should have the same ProviderType
	summaryA := summaries[0]
	model, closer, err := s.StatePool.GetModel(summaryA.UUID)
	defer closer()
	c.Assert(err, jc.ErrorIsNil)
	conf, err := model.Config()
	c.Assert(err, jc.ErrorIsNil)
	c.Check(summaryA.ProviderType, gc.Equals, conf.Type())
}

func (s *ModelSummariesSuite) TestContainsModelStatus(c *gc.C) {
	modelNameToUUID := s.Setup4Models(c)
	expectedStatus := map[string]status.StatusInfo{
		"shared": {
			Status:  status.Available,
			Message: "human message",
		},
		"user1model": {
			Status:  status.Busy,
			Message: "human message",
		},
	}
	shared, releaser, err := s.StatePool.GetModel(modelNameToUUID["shared"])
	defer releaser()
	c.Assert(err, jc.ErrorIsNil)
	err = shared.SetStatus(expectedStatus["shared"])
	user1, releaser, err := s.StatePool.GetModel(modelNameToUUID["user1model"])
	defer releaser()
	c.Assert(err, jc.ErrorIsNil)
	err = user1.SetStatus(expectedStatus["user1model"])
	c.Assert(err, jc.ErrorIsNil)
	summaries, err := s.State.ModelSummariesForUser(names.NewUserTag("user1write"))
	c.Assert(err, jc.ErrorIsNil)
	c.Check(summaries, gc.HasLen, 2)
	statuses := make(map[string]status.StatusInfo)
	for _, summary := range summaries {
		// We nil the time, because we don't want to compare it, we nil the Data map to avoid comparing an
		// empty map to a nil map
		st := summary.Status
		st.Since = nil
		st.Data = nil
		statuses[summary.Name] = st
	}
	c.Check(statuses, jc.DeepEquals, expectedStatus)
}

func (s *ModelSummariesSuite) TestContainsAccessInformation(c *gc.C) {
	modelNameToUUID := s.Setup4Models(c)
	shared, releaser, err := s.StatePool.GetModel(modelNameToUUID["shared"])
	defer releaser()
	c.Assert(err, jc.ErrorIsNil)
	err = shared.UpdateLastModelConnection(names.NewUserTag("auser"))
	s.Clock.Advance(time.Hour)
	c.Assert(err, jc.ErrorIsNil)
	timeShared := s.Clock.Now().Round(time.Second).UTC()
	err = shared.UpdateLastModelConnection(names.NewUserTag("user1write"))
	c.Assert(err, jc.ErrorIsNil)
	s.Clock.Advance(time.Hour) // give a different time for user2 accessing the shared model
	err = shared.UpdateLastModelConnection(names.NewUserTag("user2read"))
	c.Assert(err, jc.ErrorIsNil)
	user1, releaser, err := s.StatePool.GetModel(modelNameToUUID["user1model"])
	defer releaser()
	c.Assert(err, jc.ErrorIsNil)
	s.Clock.Advance(time.Hour)
	timeUser1 := s.Clock.Now().Round(time.Second).UTC()
	err = user1.UpdateLastModelConnection(names.NewUserTag("user1write"))
	c.Assert(err, jc.ErrorIsNil)

	summaries, err := s.State.ModelSummariesForUser(names.NewUserTag("user1write"))
	c.Assert(err, jc.ErrorIsNil)
	c.Check(summaries, gc.HasLen, 2)
	times := make(map[string]time.Time)
	access := make(map[string]permission.Access)
	for _, summary := range summaries {
		c.Assert(summary.UserLastConnection, gc.NotNil, gc.Commentf("nil time for %v", summary.Name))
		times[summary.Name] = summary.UserLastConnection.UTC()
		access[summary.Name] = summary.Access
	}
	c.Check(times, gc.DeepEquals, map[string]time.Time{
		"shared":     timeShared,
		"user1model": timeUser1,
	})
	c.Check(access, gc.DeepEquals, map[string]permission.Access{
		"shared":     permission.WriteAccess,
		"user1model": permission.AdminAccess,
	})
}

func (s *ModelSummariesSuite) TestContainsMachineInformation(c *gc.C) {
	modelNameToUUID := s.Setup4Models(c)
	shared, releaser, err := s.StatePool.Get(modelNameToUUID["shared"])
	defer releaser()
	c.Assert(err, jc.ErrorIsNil)
	onecore := uint64(1)
	twocores := uint64(2)
	threecores := uint64(3)
	m0, err := shared.AddMachine("quantal", state.JobHostUnits)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(m0.Life(), gc.Equals, state.Alive)
	err = m0.SetInstanceInfo("i-12345", "nonce", &instance.HardwareCharacteristics{
		CpuCores: &onecore,
	}, nil, nil, nil, nil)
	c.Assert(err, jc.ErrorIsNil)
	m1, err := shared.AddMachine("quantal", state.JobHostUnits)
	c.Assert(err, jc.ErrorIsNil)
	err = m1.SetInstanceInfo("i-45678", "nonce", &instance.HardwareCharacteristics{
		CpuCores: &twocores,
	}, nil, nil, nil, nil)
	c.Assert(err, jc.ErrorIsNil)
	m2, err := shared.AddMachine("quantal", state.JobHostUnits)
	c.Assert(err, jc.ErrorIsNil)
	err = m2.SetInstanceInfo("i-78901", "nonce", &instance.HardwareCharacteristics{
		CpuCores: &threecores,
	}, nil, nil, nil, nil)
	c.Assert(err, jc.ErrorIsNil)
	// No instance
	_, err = shared.AddMachine("quantal", state.JobHostUnits)
	c.Assert(err, jc.ErrorIsNil)
	// Dying instance, should not count to Cores or Machine count
	mDying, err := shared.AddMachine("quantal", state.JobHostUnits)
	c.Assert(err, jc.ErrorIsNil)
	err = mDying.SetInstanceInfo("i-78901", "nonce", &instance.HardwareCharacteristics{
		CpuCores: &threecores,
	}, nil, nil, nil, nil)
	c.Assert(err, jc.ErrorIsNil)
	err = mDying.Destroy()
	c.Assert(err, jc.ErrorIsNil)
	// Instance data, but no core count
	m4, err := shared.AddMachine("quantal", state.JobHostUnits)
	c.Assert(err, jc.ErrorIsNil)
	arch := "amd64"
	err = m4.SetInstanceInfo("i-78901", "nonce", &instance.HardwareCharacteristics{
		Arch: &arch,
	}, nil, nil, nil, nil)
	c.Assert(err, jc.ErrorIsNil)

	summaries, err := s.State.ModelSummariesForUser(names.NewUserTag("user1write"))
	c.Assert(err, jc.ErrorIsNil)
	c.Check(summaries, gc.HasLen, 2)
	var sharedSummary *state.ModelSummary
	for _, summary := range summaries {
		if summary.Name == "shared" {
			s := summary
			sharedSummary = &s
		}
	}
	c.Assert(sharedSummary, gc.NotNil)
	c.Check(sharedSummary.MachineCount, gc.Equals, int64(5))
	c.Check(sharedSummary.CoreCount, gc.Equals, int64(1+2+3))
}
