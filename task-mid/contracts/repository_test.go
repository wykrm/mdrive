package contracts

import "testing"

func TestPlan_AddAndMutations(t *testing.T) {
	plan := NewPlan()
	mutation := &Mutation{Table: "accounts", ID: "a1"}

	plan.Add(nil)
	plan.Add(mutation)

	mutations := plan.Mutations()
	if got, want := len(mutations), 1; got != want {
		t.Fatalf("mutation count = %d, want %d", got, want)
	}
	if mutations[0] != mutation {
		t.Fatalf("mutation = %p, want %p", mutations[0], mutation)
	}

	mutations[0] = nil
	if plan.Mutations()[0] == nil {
		t.Error("Mutations() exposed the plan's internal slice")
	}
}

func TestNilPlan_Mutations(t *testing.T) {
	var plan *Plan
	if mutations := plan.Mutations(); mutations != nil {
		t.Errorf("Mutations() = %#v, want nil", mutations)
	}
}
