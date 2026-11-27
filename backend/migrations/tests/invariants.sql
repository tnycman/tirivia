-- Schema invariants: every MUST FAIL step errors, every MUST PASS step succeeds (make db-test).

\echo '--- 1. MUST PASS: balanced transaction (pool -100, user +100)'
INSERT INTO ledger_accounts (id, owner_kind, currency, allow_negative) VALUES
 ('00000000-0000-7000-8000-000000000001','prize_pool','EUR', true),
 ('00000000-0000-7000-8000-000000000002','user','EUR', false);
BEGIN;
INSERT INTO ledger_transactions (id, idempotency_key, kind) VALUES
 ('00000000-0000-7000-8000-0000000000a1','show1:win:u2','prize_credit');
INSERT INTO ledger_entries (id, transaction_id, account_id, amount_minor, prev_balance_minor, new_balance_minor, account_version) VALUES
 ('00000000-0000-7000-8000-0000000000e1','00000000-0000-7000-8000-0000000000a1','00000000-0000-7000-8000-000000000001',-100,0,-100,1),
 ('00000000-0000-7000-8000-0000000000e2','00000000-0000-7000-8000-0000000000a1','00000000-0000-7000-8000-000000000002', 100,0, 100,1);
COMMIT;
SELECT count(*) AS committed_entries_expect_2 FROM ledger_entries;

\echo '--- 2. MUST FAIL: unbalanced transaction rejected at COMMIT'
BEGIN;
INSERT INTO ledger_transactions (id, idempotency_key, kind) VALUES
 ('00000000-0000-7000-8000-0000000000a2','bad','prize_credit');
INSERT INTO ledger_entries (id, transaction_id, account_id, amount_minor, prev_balance_minor, new_balance_minor, account_version) VALUES
 ('00000000-0000-7000-8000-0000000000e3','00000000-0000-7000-8000-0000000000a2','00000000-0000-7000-8000-000000000002', 50,100,150,2);
COMMIT;
SELECT count(*) AS entries_expect_still_2 FROM ledger_entries;

\echo '--- 3. MUST FAIL: UPDATE on ledger_entries'
UPDATE ledger_entries SET amount_minor = 999 WHERE id = '00000000-0000-7000-8000-0000000000e2';
\echo '--- 4. MUST FAIL: DELETE on ledger_entries'
DELETE FROM ledger_entries;

\echo '--- 5. MUST FAIL: duplicate ledger idempotency key'
INSERT INTO ledger_transactions (id, idempotency_key, kind) VALUES
 ('00000000-0000-7000-8000-0000000000a3','show1:win:u2','prize_credit');

\echo '--- 6. MUST FAIL: question approved by its own author (four-eyes)'
INSERT INTO questions (id, kind, difficulty, ciphertext, data_key_enc, author_id, approver_1, approver_2, status) VALUES
 ('00000000-0000-7000-8000-0000000000b1','general',2,'\x00','\x00',
  '00000000-0000-7000-8000-00000000aaaa','00000000-0000-7000-8000-00000000aaaa','00000000-0000-7000-8000-00000000bbbb','approved');

\echo '--- 7. MUST PASS: question approved by two distinct non-authors'
INSERT INTO questions (id, kind, difficulty, ciphertext, data_key_enc, author_id, approver_1, approver_2, status) VALUES
 ('00000000-0000-7000-8000-0000000000b2','general',2,'\x00','\x00',
  '00000000-0000-7000-8000-00000000aaaa','00000000-0000-7000-8000-00000000bbbb','00000000-0000-7000-8000-00000000cccc','approved');

\echo '--- 8. MUST FAIL: country eligibility change without a decision reference'
INSERT INTO country_eligibility (country_code, eligibility, decision_ref, changed_by) VALUES
 ('MK','eligible','   ','00000000-0000-7000-8000-00000000aaaa');

\echo '--- 9. MUST PASS: launch eligibility per owner decision LG-1 (all countries eligible)'
INSERT INTO country_eligibility (country_code, eligibility, decision_ref, changed_by) VALUES
 ('AL','eligible','LG-1 owner 2026-11-25','00000000-0000-7000-8000-00000000aaaa'),
 ('XK','eligible','LG-1 owner 2026-11-25','00000000-0000-7000-8000-00000000aaaa'),
 ('MK','eligible','LG-1 owner 2026-11-25','00000000-0000-7000-8000-00000000aaaa');

\echo '--- 10. MUST PASS: chat switch rows (global off, show-1 on)'
INSERT INTO feature_switches (key, scope, scope_id, enabled, changed_by) VALUES
 ('chat.enabled','global','',false,'00000000-0000-7000-8000-00000000aaaa'),
 ('chat.enabled','show','show-1',true,'00000000-0000-7000-8000-00000000aaaa');
SELECT key, scope, scope_id, enabled FROM feature_switches ORDER BY scope;

\echo '--- 11. MUST FAIL: show-scoped switch without a show id'
INSERT INTO feature_switches (key, scope, scope_id, enabled, changed_by) VALUES
 ('chat.enabled','show','',true,'00000000-0000-7000-8000-00000000aaaa');

\echo '--- 12. MUST FAIL: UPDATE on audit_log'
INSERT INTO audit_log (id, actor_id, action, entity, entity_id, hash) VALUES
 ('00000000-0000-7000-8000-0000000000c1','00000000-0000-7000-8000-00000000aaaa','switch.set','feature_switches','chat.enabled','\x01');
UPDATE audit_log SET action = 'tampered';

\echo '--- 13. MUST PASS: new user is NOT on the winners wall by default (opt-in, UF-10)'
INSERT INTO users (id, phone_hash, phone_enc, country_code) VALUES
 ('00000000-0000-7000-8000-0000000000d1','\x01','\x01','AL'),
 ('00000000-0000-7000-8000-0000000000d2','\x02','\x02','XK');
SELECT show_on_winners_wall AS winners_wall_expect_f FROM users WHERE id = '00000000-0000-7000-8000-0000000000d1';

\echo '--- 14. MUST FAIL: a late joiner recorded as winner (AC-8)'
INSERT INTO shows (id, starts_at) VALUES ('00000000-0000-7000-8000-0000000000f1', now());
INSERT INTO participations (show_id, user_id, joined_at_q, last_ok, prize_eligible, outcome) VALUES
 ('00000000-0000-7000-8000-0000000000f1','00000000-0000-7000-8000-0000000000d1', 3, 12, true, 'winner');

\echo '--- 15. MUST FAIL: a second prize-eligible account bound to the same device (AC-6)'
INSERT INTO device_prize_bindings (user_id, platform, device_ref) VALUES
 ('00000000-0000-7000-8000-0000000000d1','android','dev-ref-1');
INSERT INTO device_prize_bindings (user_id, platform, device_ref) VALUES
 ('00000000-0000-7000-8000-0000000000d2','android','dev-ref-1');
