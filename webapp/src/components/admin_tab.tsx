import React, {useState} from 'react';
import {useDispatch, useSelector} from 'react-redux';

import {executeConfirm, hideConfirm, saveDesk, showConfirm} from '../actions';
import type {GlobalState} from '../types/mattermost';
import {getPluginStateFromGlobal} from '../types/mattermost';
import type {Desk} from '../types';

import ConfirmDialog from './confirm_dialog';

const AdminTab: React.FC = () => {
    const dispatch = useDispatch();
    const {desks, locations, adminSaving, error, confirmDialog} = useSelector((state: GlobalState) => getPluginStateFromGlobal(state));
    const defaultLocationId = locations[0]?.id || '';

    const [editingId, setEditingId] = useState<string | null>(null);
    const [form, setForm] = useState({
        name: '',
        sort_order: 0,
        is_active: true,
    });

    const resetForm = () => {
        setEditingId(null);
        setForm({name: '', sort_order: desks.length, is_active: true});
    };

    const startEdit = (desk: Desk) => {
        setEditingId(desk.id);
        setForm({
            name: desk.name,
            sort_order: desk.sort_order,
            is_active: desk.is_active,
        });
    };

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        if (!form.name.trim() || !defaultLocationId) {
            return;
        }
        saveDesk(dispatch, editingId, {
            location_id: defaultLocationId,
            name: form.name.trim(),
            sort_order: form.sort_order,
            is_active: form.is_active,
        }).then(() => resetForm());
    };

    const handleDelete = (desk: Desk) => {
        dispatch(showConfirm({
            action: 'delete_desk',
            deskId: desk.id,
            deskName: desk.name,
        }));
    };

    return (
        <div className='freedesk-admin-tab'>
            <h3 className='freedesk-admin-heading'>{editingId ? 'デスク編集' : 'デスク追加'}</h3>
            <form
                className='freedesk-admin-form'
                onSubmit={handleSubmit}
            >
                <label>
                    名称
                    <input
                        type='text'
                        value={form.name}
                        onChange={(e) => setForm({...form, name: e.target.value})}
                        maxLength={64}
                        required={true}
                    />
                </label>
                <label>
                    表示順
                    <input
                        type='number'
                        value={form.sort_order}
                        onChange={(e) => setForm({...form, sort_order: Number(e.target.value)})}
                        min={0}
                    />
                </label>
                <label className='freedesk-admin-checkbox'>
                    <input
                        type='checkbox'
                        checked={form.is_active}
                        onChange={(e) => setForm({...form, is_active: e.target.checked})}
                    />
                    有効
                </label>
                <div className='freedesk-admin-form-actions'>
                    <button
                        type='submit'
                        className='btn btn-primary btn-sm'
                        disabled={adminSaving}
                    >
                        {editingId ? '更新' : '追加'}
                    </button>
                    {editingId && (
                        <button
                            type='button'
                            className='btn btn-tertiary btn-sm'
                            onClick={resetForm}
                        >
                            キャンセル
                        </button>
                    )}
                </div>
            </form>

            <table className='freedesk-admin-table'>
                <thead>
                    <tr>
                        <th>名称</th>
                        <th>表示順</th>
                        <th>状態</th>
                        <th>操作</th>
                    </tr>
                </thead>
                <tbody>
                    {desks.map((desk) => (
                        <tr key={desk.id}>
                            <td>{desk.name}</td>
                            <td>{desk.sort_order}</td>
                            <td>{desk.is_active ? '有効' : '無効'}</td>
                            <td>
                                <button
                                    type='button'
                                    className='btn btn-tertiary btn-xs'
                                    onClick={() => startEdit(desk)}
                                >
                                    編集
                                </button>
                                <button
                                    type='button'
                                    className='btn btn-tertiary btn-xs'
                                    onClick={() => handleDelete(desk)}
                                >
                                    削除
                                </button>
                            </td>
                        </tr>
                    ))}
                </tbody>
            </table>

            {error && <div className='freedesk-error'>{error}</div>}

            {confirmDialog && confirmDialog.action === 'delete_desk' && (
                <ConfirmDialog
                    dialog={confirmDialog}
                    onConfirm={() => executeConfirm(dispatch, confirmDialog, 0, 0)}
                    onCancel={() => dispatch(hideConfirm())}
                />
            )}
        </div>
    );
};

export default AdminTab;
