---
description: "Principal Engineer Examples Quality companion for examples skill."
---

# Principal Engineer Examples Quality

> Part of [agent-prompt-engineering](../INDEX.lego.md) skill examples

**Problem**: Generic code examples that don't show agent's specific workflow.

## Before (Score: 3.0/5 on Example Quality)

```markdown
<example>
user: "Implement the feature"
assistant: "I'll implement the feature for you."
</example>

## Code Patterns

Here are some general patterns:

```python
# Error handling
try:
    result = do_something()
except Exception as e:
    logger.error(f"Failed: {e}")
```
```

**Issues annotated**:
- Example says nothing about agent's actual behavior
- Code patterns are generic Python, not workflow-specific
- No TDD reference, no implementation verification

## After (Score: 4.5/5 on Example Quality)

```markdown
<example>
Context: TDD-0042 approved, specifies PaymentService with process() and refund() methods
user: "Implement TDD-0042"
assistant: "Implementing PaymentService from TDD-0042.

Reading TDD to extract:
- Component boundaries: PaymentService, PaymentGateway interface
- API contracts: process(amount, currency) -> TransactionResult
- Test requirements: Unit tests for success, failure, timeout cases

Creating implementation:

```python
# src/services/payment_service.py
class PaymentService:
    def __init__(self, gateway: PaymentGateway):
        self.gateway = gateway

    def process(self, amount: int, currency: str) -> TransactionResult:
        # Validate per TDD-0042 section 3.2
        if amount <= 0:
            raise InvalidAmountError(amount)
        ...
```

Verification:
- [ ] All TDD components implemented
- [ ] Tests pass: pytest src/tests/test_payment_service.py
- [ ] Type hints match TDD contracts"
</example>
```

## Key Improvements

- Shows actual TDD-to-implementation workflow
- References specific TDD sections
- Demonstrates verification steps
- Code is workflow-specific, not generic
