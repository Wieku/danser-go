#version 330

in vec3 in_position;
in vec3 center;
//in vec2 in_tex_coord;

uniform mat4 proj;
uniform mat4 trans;

//out vec2 tex_coord;
out float distance_inv;
void main()
{
    gl_Position = proj * ((trans * vec4(in_position-center, 1.0))+vec4(center, 0.0));
    distance_inv = 1.0f - in_position.z;
    //tex_coord = in_tex_coord;
}