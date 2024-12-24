package dev.hungq.movie_service.box;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;

import jakarta.persistence.CollectionTable;
import jakarta.persistence.Column;
import jakarta.persistence.ElementCollection;
import jakarta.persistence.Entity;
import jakarta.persistence.GeneratedValue;
import jakarta.persistence.GenerationType;
import jakarta.persistence.Id;

@Entity
public class Box {
	@Id 
	@GeneratedValue(strategy=GenerationType.IDENTITY)
	private Integer id;
	@Column(name = "owner_id")
	private Integer ownerId;
	@Column(name = "msg_box_id")
	private Integer msgBoxId;
	@Column(name = "elapsed")
	private Float elapsed;
	@Column(name = "movie_url")
	private String movieUrl;
	@Column(name = "password")
	private String password;
	
	@ElementCollection
	@CollectionTable(name = "box_user")
    @Column(name = "user_id")
	private List<Integer> userIds = new ArrayList<>();
	
	public Box() { }
	
	public Box(Integer ownerId, String password) {
		super();
		this.ownerId = ownerId;
		this.password = password;
		userIds.add(ownerId);
    }
	
	public Integer getId() {
		return id;
	}

	public void setId(Integer id) {
		this.id = id;
	}

	public Integer getOwnerId() {
		return ownerId;
	}

	public void setOwnerId(Integer ownerId) {
		this.ownerId = ownerId;
	}

	public Integer getMsgBoxId() {
		return msgBoxId;
	}

	public void setMsgBoxId(Integer msgBoxId) {
		this.msgBoxId = msgBoxId;
	}

	public Float getElapsed() {
		return elapsed;
	}

	public void setElapsed(Float elapsed) {
		this.elapsed = elapsed;
	}

	public String getMovieUrl() {
		return movieUrl;
	}

	public void setMovieUrl(String movieUrl) {
		this.movieUrl = movieUrl;
	}

	public String getPassword() {
		return password;
	}

	public void setPassword(String password) {
		this.password = password;
	}
	
	public List<Integer> getUserIds() {
		return userIds;
	}

	@Override
	public String toString() {
		return "Box [id=" + id + ", ownerId=" + ownerId + ", msgBoxId=" + msgBoxId + ", elapsed=" + elapsed
				+ ", movieUrl=" + movieUrl + ", password=" + password + "]";
	}
}